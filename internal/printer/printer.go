package printer

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"os"
	"runtime"
)

var Checkmark = "✔"
var Cross = "✖"
var GreenText = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render
var RedText = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render
var TableHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00")).Render

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

func NewTable() *table.Table {
	return table.New().Border(lipgloss.NormalBorder()).Width(160).BorderHeader(true)
}
