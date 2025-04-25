package types

import "errors"

type PRIdentifier struct {
	Owner      string
	Repository string
	PRNum      int
}

type CommentLevel uint8

const (
	Never CommentLevel = iota
	Failure
	Always
)

func (l CommentLevel) AtLeast(baseline CommentLevel) bool {
	return l >= baseline
}

func CommentLevelFromStr(str string) (CommentLevel, error) {
	switch str {
	case "never":
		return Never, nil
	case "failure":
		return Failure, nil
	case "always":
		return Always, nil
	}

	return 0, errors.New(`value must be in ["never", "failure", "always"]`)
}

type Config struct {
	AuthToken       string
	TriggeringActor string

	PR PRIdentifier

	Merge        bool
	CommentLevel CommentLevel
	AllowedRoles []string
}

type PRInfo struct {
	BaseBranch string
	TargetSHA  string
}
