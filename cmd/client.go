package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/PrajnaAvidya/yapatype/client"
	"github.com/PrajnaAvidya/yapatype/config"
	"github.com/PrajnaAvidya/yapatype/executor"
	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "start a client that executes commands from the server",
	RunE:  runClient,
}

var (
	clientName       string
	clientServer     string
	clientExecutor   string
	clientKittySocket string
	clientConfig     string
)

func init() {
	clientCmd.Flags().StringVarP(&clientName, "name", "n", "", "client name for identification (default: hostname-executor)")
	clientCmd.Flags().StringVarP(&clientServer, "server", "s", "", "server websocket url (default: ws://localhost:9999)")
	clientCmd.Flags().StringVarP(&clientExecutor, "executor", "e", "", "input method: auto, ydotool, kitty, osascript")
	clientCmd.Flags().StringVarP(&clientKittySocket, "kitty-socket", "k", "", "kitty socket path (implies --executor kitty)")
	clientCmd.Flags().StringVarP(&clientConfig, "config", "c", "", "config file path")

	rootCmd.AddCommand(clientCmd)
}

func runClient(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(clientConfig)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// cli flags override config
	if clientName != "" {
		cfg.Client.Name = clientName
	}
	if clientServer != "" {
		cfg.Client.ServerURL = clientServer
	}
	if clientExecutor != "" {
		cfg.Client.Executor = clientExecutor
	}
	if clientKittySocket != "" {
		cfg.Client.KittySocket = &clientKittySocket
		// kitty socket implies kitty executor
		if cfg.Client.Executor == "auto" {
			cfg.Client.Executor = "kitty"
		}
	}

	fmt.Printf("yapatype client v%s\n", version)
	fmt.Printf("name: %s\n", cfg.Client.Name)

	// create executor
	exec, err := executor.NewExecutor(&cfg.Client)
	if err != nil {
		return fmt.Errorf("create executor: %w", err)
	}
	fmt.Printf("executor: %s\n", exec.Name())

	// create client
	c := client.New(
		cfg.Client.ServerURL,
		cfg.Client.Name,
		runtime.GOOS,
		exec,
	)

	// load .whisper-prompt
	prompt := loadPromptFile()
	if prompt != "" {
		fmt.Printf("loaded .whisper-prompt (%d chars)\n", len(prompt))
	}

	// signal handling
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nshutting down...")
		cancel()
		c.Stop()
	}()

	// send prompt after first connection
	if prompt != "" {
		go func() {
			// wait briefly for connection to establish
			time.Sleep(500 * time.Millisecond)
			if err := c.SendPrompt(ctx, prompt); err != nil {
				// not critical, just log it
				fmt.Printf("failed to send prompt: %v\n", err)
			}
		}()
	}

	return c.Run(ctx)
}

// loadPromptFile reads .whisper-prompt from current directory
func loadPromptFile() string {
	data, err := os.ReadFile(".whisper-prompt")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
