package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func CallPythonToS() bool {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("python", "tos.py")
	} else {
		cmd = exec.Command("python3", "tos.py")
	}

	// Connect stdin/stdout so Python can interact with user
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running Python ToS: %v\n", err)
		return false
	}

	// Python script should exit with code 0 for acceptance, 1 for rejection
	return cmd.ProcessState.ExitCode() == 0
}
