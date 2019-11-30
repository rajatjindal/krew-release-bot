package krew

import (
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
