// Package cmd, This file defines the root command for the GitServe CLI application.
// All other commands (run, list, stop, init) will be attached to this root command.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	// Viper will be used later, likely initialized via an initConfig function
	// "github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands.
// It provides the main help text and application description.
var rootCmd = &cobra.Command{
	Use:   "gitserve",
	Short: "GitServe runs commands (and dev servers) from isolated Git branches.",
	Long: `GitServe helps manage development workflows by allowing you to easily
run commands against specific Git branches in isolated environments.
How?
It clones the desired branch, runs setup commands (like dependency installs),
and executes your primary command (like starting a dev server).
It can manage these processes running in the background.`,
	// Uncomment SilenceUsage if you don't want Cobra to print usage info on error.
	// The error itself will still be printed.
	// SilenceUsage: true,
}

// Execute is called by main.main(). It's the main entry point for Cobra execution.
// It parses the command line arguments and runs the appropriate command logic.
func Execute() {
	// Set output for errors to stderr
	rootCmd.SetErr(os.Stderr)
	// Execute the root command. Cobra handles parsing args and running subcommands.
	if err := rootCmd.Execute(); err != nil {
		// Cobra usually prints the error, so we just exit with a non-zero status.
		os.Exit(1)
	}
}
