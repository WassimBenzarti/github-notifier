package terminal

import "fmt"

var (
	Reset     = "\033[0m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	Gray      = "\033[37m"
	White     = "\033[97m"
	LightBlue = "\033[94m"
)

func ColorfulPrintf(color string, str string, args ...any) {
	fmt.Printf("%s%s%s", color, fmt.Sprintf(str, args...), Reset)
}
