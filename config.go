package fileshare

import (
	"flag"
	"fmt"
	"os"
	"os/user"
)

const defaultPort = 8000

type Config struct {
	TargetDir   string
	Port        int
	Host        string
	AutoOpen    bool
	ShowVersion bool
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func parseConfig() (Config, error) {
	cfg := Config{Host: "0.0.0.0"}
	flag.StringVar(&cfg.TargetDir, "t", "", "Directory to share (default: current dir, use 'home' for home directory)")
	flag.IntVar(&cfg.Port, "p", defaultPort, "Port to run server on")
	flag.StringVar(&cfg.Host, "host", cfg.Host, "Host/IP to listen on")
	flag.BoolVar(&cfg.AutoOpen, "open", true, "Automatically open browser")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "Print version and exit")
	flag.Parse()

	if cfg.ShowVersion {
		return cfg, nil
	}

	if cfg.TargetDir == "home" {
		currentUser, err := user.Current()
		if err != nil {
			return Config{}, fmt.Errorf("failed to get current user: %w", err)
		}
		cfg.TargetDir = currentUser.HomeDir
	} else if cfg.TargetDir == "" {
		defaultDir, err := os.Getwd()
		if err != nil {
			return Config{}, fmt.Errorf("failed to get working directory: %w", err)
		}
		cfg.TargetDir = defaultDir
	}
	return cfg, nil
}
