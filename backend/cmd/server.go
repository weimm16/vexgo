package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// Config holds the server configuration from command line arguments and/or config file
type Config struct {
	Addr    string // Address to listen on (e.g., "0.0.0.0" or "127.0.0.1")
	Port    int    // Port to listen on
	DataDir string // Data directory for storing sqlite database and media files

	// Database configuration
	DBType     string `yaml:"db_type"`     // Database type: "sqlite" or "mysql"
	DBHost     string `yaml:"db_host"`     // MySQL host (for mysql)
	DBPort     int    `yaml:"db_port"`     // MySQL port (for mysql)
	DBUser     string `yaml:"db_user"`     // MySQL user (for mysql)
	DBPassword string `yaml:"db_password"` // MySQL password (for mysql)
	DBName     string `yaml:"db_name"`     // MySQL database name (for mysql)
}

// ParseFlags parses command line flags and returns the server configuration
func ParseFlags() *Config {
	configFile := flag.String("c", "", "Path to configuration file (YAML format)")
	addr := flag.String("addr", "0.0.0.0", "Address to listen on (default: 0.0.0.0)")
	port := flag.Int("port", 3001, "Port to listen on (default: 3001)")
	dataDir := flag.String("data", "./data", "Data directory for storing sqlite database and media files (default: ./data)")

	// Parse command line flags
	flag.Parse()

	// Default configuration
	cfg := &Config{
		Addr:    *addr,
		Port:    *port,
		DataDir: *dataDir,
	}

	// If config file is specified, load it
	if *configFile != "" {
		if err := loadConfigFile(*configFile, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config file %s: %v\n", *configFile, err)
		} else {
			fmt.Printf("Loaded configuration from %s\n", *configFile)
		}
	}

	return cfg
}

// loadConfigFile loads configuration from a YAML file
func loadConfigFile(filename string, cfg *Config) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal YAML into the config struct
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return nil
}

// GetListenAddr returns the full listen address in the format "addr:port"
func (c *Config) GetListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Addr, c.Port)
}

// PrintUsage prints usage information for the server command
func PrintUsage() {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
}
