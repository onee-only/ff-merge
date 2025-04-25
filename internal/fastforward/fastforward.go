package fastforward

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/onee-only/ff-merge/internal/types"
	"github.com/pkg/errors"
)

func Do(
	ctx context.Context,
	authToken string,
	prID types.PRIdentifier,
	info types.PRInfo,
	push bool,
) (success bool, err error) {
	if err := os.Mkdir("/tmp", 0755); err != nil {
		return false, errors.Wrap(err, "creating /tmp")
	}

	auth := &http.BasicAuth{
		Username: "ff-merge", // Doc says it can be anything but empty.
		Password: authToken,
	}

	baseBranchRef := plumbing.NewBranchReferenceName(info.BaseBranch)

	r, err := git.PlainCloneContext(ctx, "/tmp", true, &git.CloneOptions{
		URL: fmt.Sprintf("https://github.com/%s/%s", prID.Owner, prID.Repository),

		Auth:              auth,
		ReferenceName:     baseBranchRef,
		ShallowSubmodules: true,
	})

	if err != nil {
		return false, errors.Wrap(err, "cloning repository")
	}

	ref := plumbing.NewHashReference("don't care", plumbing.NewHash(info.TargetSHA))

	err = r.Merge(*ref, git.MergeOptions{Strategy: git.FastForwardMerge})
	if err != nil {
		if errors.Is(err, git.ErrFastForwardMergeNotPossible) {
			err = nil
		}
		return false, errors.Wrap(err, "merging branches")
	}

	if !push {
		return true, nil
	}

	err = r.PushContext(ctx, &git.PushOptions{
		RefSpecs: []config.RefSpec{
			// Push head to intended branch.
			config.RefSpec(fmt.Sprintf("HEAD:%s", baseBranchRef)),
		},
		Auth: auth,
	})
	if err != nil {
		return false, errors.Wrap(err, "pushing branch")
	}

	return true, nil
}
