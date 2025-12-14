package cmd

import (
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "yapatype",
	Short:   "voice-to-terminal tool with server/client architecture",
	Long:    "yapatype is a voice-to-terminal tool that captures audio, transcribes it, and executes commands across multiple clients.",
	Version: version,
}

func init() {
	// use -V as shorthand for --version (cobra default is -v)
	rootCmd.Flags().BoolP("version", "V", false, "show version")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
