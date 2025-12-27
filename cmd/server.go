package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/PrajnaAvidya/yapatype/config"
	"github.com/PrajnaAvidya/yapatype/server"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start the voice capture and transcription server",
	RunE:  runServer,
}

var (
	serverPort    int
	serverHost    string
	serverConfig  string
	serverNoSound bool
	serverMic     string
)

func init() {
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 0, "websocket server port (default: 9999)")
	serverCmd.Flags().StringVar(&serverHost, "host", "", "websocket server host (default: 0.0.0.0)")
	serverCmd.Flags().StringVarP(&serverConfig, "config", "c", "", "config file path")
	serverCmd.Flags().BoolVar(&serverNoSound, "no-sound", false, "disable audio feedback sounds")
	serverCmd.Flags().StringVar(&serverMic, "mic", "", "microphone to use (substring match)")
	// alias for --mic
	serverCmd.Flags().StringVar(&serverMic, "microphone", "", "microphone to use (substring match)")
	serverCmd.Flags().MarkHidden("microphone")

	rootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(serverConfig)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// cli flags override config
	if cmd.Flags().Changed("port") {
		cfg.Server.Port = serverPort
	}
	if cmd.Flags().Changed("host") {
		cfg.Server.Host = serverHost
	}
	if serverNoSound {
		cfg.Server.Sounds.Enabled = false
	}
	if serverMic != "" {
		cfg.Server.Microphone = &serverMic
	}

	fmt.Printf("yapatype server v%s\n", version)

	// create server
	srv := server.NewServer(cfg.Server.Host, cfg.Server.Port, cfg.Server.Aliases)

	// signal handling
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nshutting down...")
		cancel()
		srv.Stop()
	}()

	return srv.Run(ctx)
}
