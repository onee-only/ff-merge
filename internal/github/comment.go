package github

import (
	"context"

	gh "github.com/google/go-github/v71/github"
	"github.com/onee-only/ff-merge/internal/types"
	"github.com/pkg/errors"
)

func Comment(
	ctx context.Context,
	client *gh.Client,
	prID types.PRIdentifier,
	message string,
) error {
	comment := &gh.PullRequestComment{Body: &message}

	_, _, err := client.PullRequests.CreateComment(
		ctx,
		prID.Owner, prID.Repository, prID.PRNum,
		comment,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create comment")
	}

	return nil
}
