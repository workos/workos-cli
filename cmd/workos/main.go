package main

import (
	"github.com/workos/workos-cli/internal/cmd"
)

func main() {
	cmd.SetVersion("dev")
	cmd.Execute()
}
