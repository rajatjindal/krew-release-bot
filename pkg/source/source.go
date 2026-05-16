package source

import (
	"net/http"

	"github.com/rajatjindal/krew-release-bot/pkg/types"
)

// Source is a release source interface
type Source interface {
	Parse(r *http.Request) (*types.ReleaseRequest, error)
}
