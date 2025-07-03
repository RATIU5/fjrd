package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/RATIU5/fjrd/internal/config"
	"github.com/RATIU5/fjrd/internal/interaction"
	"github.com/RATIU5/fjrd/internal/logger"
)

func main() {
	var (
		logLevel  = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		logFormat = flag.String("log-format", "text", "Log format (text, json)")
		timeout   = flag.Duration("timeout", 30*time.Second, "Operation timeout")
		quiet     = flag.Bool("quiet", false, "Suppress non-error output")
		verbose   = flag.Bool("verbose", false, "Enable verbose logging (equivalent to -log-level=debug)")
		help      = flag.Bool("help", false, "Show help message")
	)

	const appName string = "fjrd"

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <config-path>\n\n", appName)
		fmt.Fprintf(os.Stderr, "A macOS configuration management tool that applies system settings via TOML files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s config.toml\n", appName)
		fmt.Fprintf(os.Stderr, "  %s -verbose owner/repo\n", appName)
		fmt.Fprintf(os.Stderr, "  %s -log-level=debug https://example.com/config.toml\n", appName)
		fmt.Fprintf(os.Stderr, "  %s -quiet -timeout=60s config.toml\n", appName)
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: config-path is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	level := logger.ParseLevel(*logLevel)
	if *verbose {
		level = logger.LevelDebug
	}
	if *quiet {
		level = logger.LevelError
	}

	var log *logger.Logger
	if *logFormat == "json" {
		log = logger.NewJSON(level, os.Stderr)
	} else {
		log = logger.New(level, os.Stderr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	configPath := flag.Arg(0)

	log.Debug("Starting fjrd", "config_path", configPath, "timeout", *timeout)

	cfg, err := config.LoadConfig(ctx, configPath, log)
	if err != nil {
		log.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	log.Debug("Configuration loaded successfully")

	// Check if raw defaults require user approval
	if cfg.RequiresRawDefaultsApproval() && !*quiet {
		log.Debug("Raw defaults detected, requesting user approval")
		commands := cfg.ListRawDefaults()
		approved, err := interaction.GetUserApproval(commands)
		if err != nil {
			log.Error("Failed to get user approval", "error", err)
			os.Exit(1)
		}
		if !approved {
			log.Info("Operation cancelled by user")
			os.Exit(0)
		}
		log.Debug("User approved raw defaults execution")
	}

	if err := cfg.Execute(ctx, log); err != nil {
		log.Error("Failed to execute config", "error", err)
		os.Exit(1)
	}

	if !*quiet {
		log.Info("Configuration applied successfully")
	}
}
