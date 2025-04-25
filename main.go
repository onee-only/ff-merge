package main

import (
	"context"
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

	conf.AuthToken = actions.GetInput("GITHUB_TOKEN")

	level, err := types.CommentLevelFromStr(actions.GetInput("comment"))
	if err != nil {
		return types.Config{}, errors.Wrap(err, "failed to get a value of comment")
	}
	conf.CommentLevel = level

	mergeStr := actions.GetInput("merge")
	switch mergeStr {
	case "true":
		conf.Merge = true
	case "false":
		conf.Merge = false
	default:
		return types.Config{}, errors.New(`merge must be "true" or "false"`)
	}

	return conf, nil
}
