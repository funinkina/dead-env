package history

import "errors"

var (
	ErrInvalidHistoryDir  = errors.New("invalid history directory")
	ErrGitNotFound        = errors.New("git is not available in the path")
	ErrGitCommandFailed   = errors.New("something went wrong with git command")
	ErrInvalidProfileName = errors.New("invalid profile name")
)
