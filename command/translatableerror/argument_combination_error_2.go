package translatableerror

import "strings"

// ArgumentCombinationError2 represent an error caused by using two command line
// arguments that cannot be used together.
type ArgumentCombinationError2 struct {
	Args []string
}

func (ArgumentCombinationError2) DisplayUsage() {}

func (ArgumentCombinationError2) Error() string {
	return "Incorrect Usage: The following arguments cannot be used together: {{.Args}}"
}

func (e ArgumentCombinationError2) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Args": strings.Join(e.Args, ", "),
	})
}
