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
	fmt.Printf("connecting to %s\n", cfg.Client.ServerURL)

	// TODO: run client
	return nil
}
