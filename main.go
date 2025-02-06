// Package main is only the entry point for the application.
// No other source code should be in the main package otherwise it will get
// mixed with non source code files at the root of the project.
package main

import "mark/cmd"

func main() {
	cmd.Execute()
}
