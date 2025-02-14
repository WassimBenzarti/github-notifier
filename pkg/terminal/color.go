package terminal

import "fmt"

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"
var LightBlue = "\033[94m"

func ColorfulPrintf(color string, str string, args ...any) {
	fmt.Printf("%s%s%s", color, fmt.Sprintf(str, args...), Reset)
}
