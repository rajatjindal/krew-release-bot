package krew

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluginFileName(t *testing.T) {
	pluginName := "whoami"
	expectedFileName := "whoami.yaml"
	fileName := PluginFileName(pluginName)

	assert.Equal(t, expectedFileName, fileName)
}

func TestGetPluginName(t *testing.T) {
	testcases := []struct {
		name          string
		file          string
		expectedName  string
		expectedError string
	}{
		{
			name:          "valid plugin file",
			file:          "data/valid-file.yaml",
			expectedName:  "whoami",
			expectedError: "",
		},
		{
			name:          "file does not exist",
			file:          "data/file-dont-exist.yaml",
			expectedName:  "",
			expectedError: "open data/file-dont-exist.yaml: no such file or directory",
		},
		{
			name:          "invalid plugin file",
			file:          "data/invalid-plugin-file.yaml",
			expectedName:  "",
			expectedError: "error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type index.Plugin",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			pluginName, err := GetPluginName(tc.file)
			assert.Equal(t, tc.expectedName, pluginName)

			if tc.expectedError != "" {
				assert.NotNil(t, err)
				if err != nil {
					assert.Equal(t, tc.expectedError, err.Error())
				}
			}

			if tc.expectedError == "" {
				assert.Nil(t, err)
			}
		})
	}
}
