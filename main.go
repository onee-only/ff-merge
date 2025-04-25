package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	github_sdk "github.com/google/go-github/v71/github"
	"github.com/onee-only/ff-merge/internal/fastforward"
	"github.com/onee-only/ff-merge/internal/github"
	"github.com/onee-only/ff-merge/internal/types"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-githubactions"
)

func main() {
	ctx := context.Background()
	actions := githubactions.New()

	if err := run(ctx, actions); err != nil {
		actions.Fatalf("error while running: %s", err.Error())
	}
}

func run(ctx context.Context, actions *githubactions.Action) error {
	conf, err := readConf(actions)
	if err != nil {
		return errors.Wrap(err, "failed to read configuration")
	}

	client := github_sdk.NewClient(http.DefaultClient).WithAuthToken(conf.AuthToken)

	err = github.CheckPermission(
		ctx, client,
		conf.PR, conf.TriggeringActor, conf.AllowedRoles,
	)
	if err != nil {
		if errors.Is(err, github.ErrUserNotAllowed) && conf.CommentLevel.AtLeast(types.Failure) {
			message := fmt.Sprintf(
				"@%s doesn't have enough permission to trigger fast-forward",
				conf.TriggeringActor,
			)

			if err := github.Comment(ctx, client, conf.PR, message); err != nil {
				return errors.Wrap(err, "failed to create a comment")
			}

			return nil
		}
		return errors.Wrap(err, "error while checking permission")
	}

	info, err := github.GetPRInfo(ctx, client, conf.PR)
	if err != nil {
		return errors.Wrap(err, "getting PR info")
	}

	success, err := fastforward.Do(ctx, conf.AuthToken, conf.PR, info, conf.Merge)
	if err != nil {
		return errors.Wrap(err, "unexpected error while fast forward")
	}

	var message string
	var canComment bool
	if !success && conf.CommentLevel.AtLeast(types.Failure) {
		canComment = true
		message = "cannot fast-forward `%s` to `%s`"
	}
	if success && conf.CommentLevel.AtLeast(types.Always) {
		canComment = true
		if conf.Merge {
			message = "successfully fast-forwarded `%s` to `%s`"
		} else {
			message = "available to fast-forward `%s` to `%s`"
		}
	}

	if canComment {
		message = fmt.Sprintf(message, info.BaseBranch, info.TargetSHA)
		if err := github.Comment(ctx, client, conf.PR, message); err != nil {
			return errors.Wrap(err, "failed to create a comment")
		}
	}

	return nil
}

func readConf(actions *githubactions.Action) (conf types.Config, err error) {
	context, err := actions.Context()
	if err != nil {
		return types.Config{}, errors.Wrap(err, "reading action context")
	}

	conf.PR.Owner = context.RepositoryOwner
	conf.PR.Repository = context.Repository
	conf.TriggeringActor = context.TriggeringActor

	conf.PR.PRNum, err = processEvent(context.Event)
	if err != nil {
		return types.Config{}, errors.Wrap(err, "processing event")
	}

	conf.AuthToken = actions.GetInput("GITHUB_TOKEN")

	conf.CommentLevel, err = types.CommentLevelFromStr(actions.GetInput("comment"))
	if err != nil {
		return types.Config{}, errors.Wrap(err, "failed to get a value of comment")
	}

	mergeStr := actions.GetInput("merge")
	switch mergeStr {
	case "true":
		conf.Merge = true
	case "false":
		conf.Merge = false
	default:
		return types.Config{}, errors.New(`merge must be "true" or "false"`)
	}

	allowedRolesBytes := []byte(actions.GetInput("allowed-roles"))
	if err := json.Unmarshal(allowedRolesBytes, &conf.AllowedRoles); err != nil {
		return types.Config{}, errors.Wrap(err, "failed to parse allowed-roles")
	}

	return conf, nil
}

func processEvent(event map[string]any) (prNum int, err error) {
	// Get PR number from issue.
	if issue, ok := event["issue"].(map[string]any); ok {
		if pr, ok := issue["pull_request"].(map[string]any); !ok || len(pr) == 0 {
			// The pull_request field could be {}.
			// Which means that it will be unmarshaled into empty map.
			return 0, errors.New("event is not from pull request")
		}

		num, ok := issue["number"].(int)
		if !ok {
			return 0, errors.New("event.issue.number is invalid")
		}

		return num, nil
	}

	// Get PR number from pull_request.
	if pull_request, ok := event["pull_request"].(map[string]any); ok {
		num, ok := pull_request["number"].(int)
		if !ok {
			return 0, errors.New("event.pull_request.number is invalid")
		}

		return num, nil
	}

	return 0, errors.New("event doesn't contain neither issue nor pull_request")
}
