package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/muesli/termenv"
)

var Purple = termenv.ColorProfile().Color("#6310FF")
var Red = termenv.ColorProfile().Color("#FF0000")
var Green = termenv.ColorProfile().Color("#00FF00")
var Checkmark = "✔"
var Cross = "✖"

func init() {
	if runtime.GOOS == "windows" {
		Checkmark = "√"
		Cross = "×"
	}
}

func PrintJson(val any) {
	bytes, err := json.MarshalIndent(val, "", "    ")
	if err != nil {
		PrintErrAndExit(err.Error())
	}
	fmt.Printf("%s\n", string(bytes))
}

func PrintErrAndExit(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, "Error:", msg)
	os.Exit(1)
}
