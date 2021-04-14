package cli

import "fmt"

// might add more in the future
var (
	yellow = color("\033[1;33m%s\033[0m")
	dimmed = color("\033[2;37m%s\033[0m")
)

// source: https://gist.github.com/ik5/d8ecde700972d4378d87#gistcomment-3074524
func color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}
