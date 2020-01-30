package krew

import (
	"bytes"
	"fmt"
	"strings"

	"sigs.k8s.io/krew/pkg/index/indexscanner"
	"sigs.k8s.io/krew/pkg/index/validation"
)

//ValidateOwnership validates the ownership of the plugin
func ValidateOwnership(file, expectedOwner string) error {
	if expectedOwner == "" {
		return fmt.Errorf("expectedOwner cannot be empty string")
	}

	plugin, err := indexscanner.ReadPluginFile(file)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(plugin.Spec.Homepage, fmt.Sprintf("https://github.com/%s/", expectedOwner)) {
		return fmt.Errorf("plugin homepage %s does not have prefix %s", plugin.Spec.Homepage, fmt.Sprintf("https://github.com/%s/", expectedOwner))
	}

	return nil
}

//ValidatePlugin validates the plugin spec
func ValidatePlugin(name, file string) error {
	plugin, err := indexscanner.ReadPluginFile(file)
	if err != nil {
		return err
	}

	return validation.ValidatePlugin(name, plugin)
}

//GetPluginName gets the plugin name from template .krew.yaml file
func GetPluginName(spec []byte) (string, error) {
	plugin, err := indexscanner.DecodePluginFile(bytes.NewReader(spec))
	if err != nil {
		return "", err
	}

	return plugin.GetName(), nil
}

//PluginFileName returns the plugin file with extension
func PluginFileName(name string) string {
	return fmt.Sprintf("%s%s", name, ".yaml")
}
