package krew

import (
	"io/ioutil"
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
			name:          "invalid plugin file",
			file:          "data/invalid-plugin-file.yaml",
			expectedName:  "",
			expectedError: "error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type index.Plugin",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := ioutil.ReadFile(tc.file)
			pluginName, err := GetPluginName(spec)
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

func TestValidateOwnership(t *testing.T) {
	testcases := []struct {
		name          string
		file          string
		owner         string
		expectedError string
	}{
		{
			name:          "valid plugin file, owner is empty",
			file:          "data/valid-file.yaml",
			owner:         "",
			expectedError: "expectedOwner cannot be empty string",
		},
		{
			name:          "valid plugin file, owner is incorrect",
			file:          "data/valid-file.yaml",
			owner:         "foo-bar",
			expectedError: "plugin homepage https://github.com/rajatjindal/kubectl-whoami does not have prefix https://github.com/foo-bar/",
		},
		{
			name:          "valid plugin file, owner is substring of actual owner",
			file:          "data/valid-file.yaml",
			owner:         "rajatjin",
			expectedError: "plugin homepage https://github.com/rajatjindal/kubectl-whoami does not have prefix https://github.com/rajatjin/",
		},
		{
			name:  "valid plugin file, owner is correct",
			file:  "data/valid-file.yaml",
			owner: "rajatjindal",
		},
		{
			name:          "file does not exist",
			file:          "data/file-dont-exist.yaml",
			owner:         "rajatjindal",
			expectedError: "open data/file-dont-exist.yaml: no such file or directory",
		},
		{
			name:          "invalid plugin file",
			file:          "data/invalid-plugin-file.yaml",
			owner:         "rajatjindal",
			expectedError: "error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type index.Plugin",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOwnership(tc.file, tc.owner)

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
