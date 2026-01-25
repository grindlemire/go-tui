package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/grindlemire/go-tui/pkg/lsp"
)

func runLSP(args []string) error {
	fs := flag.NewFlagSet("lsp", flag.ExitOnError)
	logPath := fs.String("log", "", "Path to log file for debugging")

	if err := fs.Parse(args); err != nil {
		return err
	}

	server := lsp.NewServer(os.Stdin, os.Stdout)

	// Set up logging if requested
	if *logPath != "" {
		logFile, err := os.OpenFile(*logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("opening log file: %w", err)
		}
		defer logFile.Close()
		server.SetLogFile(logFile)
	}

	return server.Run(context.Background())
}
