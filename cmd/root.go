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

// init() runs before main() and is used to set up flags and configurations.
func init() {
	// Call a function to initialize configuration loading (using Viper).
	// This function will likely live in internal/config but might be triggered here.
	// cobra.OnInitialize(initConfig)

	// Example: Define global persistent flags here if needed
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .gitserve.json)")
}

/*
// initConfig reads in config file and ENV variables if set.
// This function's logic will primarily be implemented in internal/config
// and called from the init() function or using cobra.OnInitialize().
func initConfig() {
    // Example Viper setup (details TBD in internal/config/config.go)
    // viper.SetConfigName(".gitserve") // Config file name without extension
    // viper.SetConfigType("json")     // Specify JSON type
    // viper.AddConfigPath(".")        // Look for config in the current directory
    // viper.AddConfigPath("$HOME/.gitserve") // Optionally look in home directory
    // viper.AutomaticEnv()            // Read in environment variables that match

    // If a config file is found, read it in.
    // if err := viper.ReadInConfig(); err == nil {
    //  fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
    // } else {
    //  // Handle cases where config file is not found, maybe it's okay?
    //  // Or only error out if a specific config file flag was passed but not found.
    // }
}
*/
