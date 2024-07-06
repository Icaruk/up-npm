package cli

import (
	"github.com/AlecAivazis/survey/v2"
)

var YesNoPromptOptions = struct {
	Yes string
	No  string
}{
	Yes: "Yes",
	No:  "No",
}

func PromptYesNo(message string) (string, error) {

	response := ""
	prompt := &survey.Select{
		Message: message,
		Options: []string{
			YesNoPromptOptions.Yes,
			YesNoPromptOptions.No,
		},
	}
	err := survey.AskOne(prompt, &response)

	return response, err
}
