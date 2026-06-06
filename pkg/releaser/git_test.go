package releaser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func TestGetBranchReferenceName(t *testing.T) {
	assert.Equal(t, plumbing.ReferenceName("refs/heads/main"), getBranchReferenceName("main"))
	assert.Equal(t, plumbing.ReferenceName("refs/heads/release"), getBranchReferenceName("refs/heads/release"))
}
