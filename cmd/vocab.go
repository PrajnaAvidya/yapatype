package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var vocabCmd = &cobra.Command{
	Use:   "vocab [project_path]",
	Short: "generate project-specific whisper vocabulary",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runVocab,
}

func init() {
	rootCmd.AddCommand(vocabCmd)
}

func runVocab(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	// resolve to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// check path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", absPath)
	}

	fmt.Printf("generating vocabulary for: %s\n", absPath)

	// TODO: generate vocab
	return nil
}
