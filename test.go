package main

import (
	"os/exec"
)

func main() {
	// This is a placeholder for the main function.
	userCode := "#include<stdio<"
	feedback := ""
	if userCode != "" {
		cmd := exec.Command("python3", "glm.py", userCode)
		output, err := cmd.Output()
		if err != nil {
			feedback = "Error executing code: " + err.Error()
		} else {
			feedback = string(output)
		}
	}
	println(feedback)
}
