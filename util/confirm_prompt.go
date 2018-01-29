package util

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

// Returns true if user selects 'Yes'. false if 'No'
func PromptForConfirmation(label, selected string) (bool, error) {
	confirmOptions := []struct {
		Name  string
		Value bool
	}{
		{"Yes", true},
		{"No", false},
	}
	confirmPrompt := promptui.Select{
		Label: label,
		Items: confirmOptions,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
			Inactive: "  {{.Name}}",
			Selected: fmt.Sprintf("  %s? {{.Name}}", selected),
		},
	}

	i, _, err := confirmPrompt.Run()
	if err != nil {
		return false, err
	}

	return confirmOptions[i].Value, nil
}
