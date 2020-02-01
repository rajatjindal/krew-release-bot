package source

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixIndent(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		indent int
		output string
	}{
		{
			name: "indent with 4 spaces",
			input: `uri: some-secure-uri
sha256: some-sha256`,
			indent: 4,
			output: `uri: some-secure-uri
    sha256: some-sha256`,
		},
		{
			name: "fix legacy 4 space indent with 6 spaces",
			input: `uri: some-secure-uri
    sha256: some-sha256`,
			indent: 6,
			output: `uri: some-secure-uri
      sha256: some-sha256`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output := indent(tc.indent, tc.input)
			assert.Equal(t, tc.output, output)
		})
	}
}
