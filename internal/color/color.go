package color

import "runtime"

var Reset = "\033[0m"
var Black = "\033[30m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var White = "\033[37m"
var Gray = "\033[90m"
var BrightRed = "\033[91m"
var BrightGreen = "\033[92m"
var BrightYellow = "\033[93m"
var BrightBlue = "\033[94m"
var BrightPurple = "\033[95m"
var BrightCyan = "\033[96m"
var BrightWhite = "\033[97m"

func init() {
	if runtime.GOOS == "windows" {
		Reset = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Purple = ""
		Cyan = ""
		White = ""
		Gray = ""
		BrightRed = ""
		BrightGreen = ""
		BrightYellow = ""
		BrightBlue = ""
		BrightPurple = ""
		BrightCyan = ""
		BrightWhite = ""
	}
}
