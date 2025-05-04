package main

import "gitserve/cmd"

func main() {
	// All the application logic starts by executing the root command
	// defined in the cmd package. Cobra handles the rest.
	cmd.Execute()
}
