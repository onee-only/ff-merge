package github

import (
	"context"
	"net/http"

	gh "github.com/google/go-github/v71/github"
	"github.com/onee-only/ff-merge/internal/types"
	"github.com/pkg/errors"
)

var ErrUserNotAllowed = errors.New("user is not allowed to run fast forward")

func CheckPermission(
	ctx context.Context,
	client *gh.Client,
	pr types.PRIdentifier,
	user string,
	allowedRoles []string,
) error {
	if len(allowedRoles) == 0 {
		// Anyone can run this.
		return nil
	}

	level, res, err := client.Repositories.GetPermissionLevel(
		ctx, pr.Owner, pr.Repository, user,
	)
	if res.StatusCode == http.StatusNotFound {
		return errors.Wrap(err, "user is not a collaborator of this repository")
	}
	if err != nil {
		return errors.Wrap(err, "getting permission level")
	}

	role := level.GetRoleName()
	for _, expected := range allowedRoles {
		if role == expected {
			return nil
		}
	}

	return errors.Wrap(ErrUserNotAllowed, "user doesn't have matching role")
}
