package github

import (
	"context"

	gh "github.com/google/go-github/v71/github"
	"github.com/onee-only/ff-merge/internal/types"
)

func GetPRInfo(
	ctx context.Context,
	client *gh.Client,
	prID types.PRIdentifier,
) (prInfo types.PRInfo, _ error) {
	pr, _, err := client.PullRequests.Get(
		ctx, prID.Owner, prID.Repository, prID.PRNum,
	)
	if err != nil {

	}

	sha := pr.GetHead().GetSHA()
	base := pr.GetBase().GetRef()

	return types.PRInfo{
		BaseBranch: base,
		TargetSHA:  sha,
	}, nil
}
