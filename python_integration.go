package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

// CallPythonToS executes the Python ToS script and returns acceptance status
func CallPythonToS() bool {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("python", "tos.py")
	} else {
		cmd = exec.Command("python3", "tos.py")
	}

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running Python ToS: %v\n", err)
		return false
	}

	// Python script should exit with code 0 for acceptance, 1 for rejection
	return cmd.ProcessState.ExitCode() == 0
}
