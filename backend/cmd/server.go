package cmd

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/goccy/go-yaml"
)

// Config holds the server configuration from command line arguments and/or config file
type Config struct {
	Addr      string // Address to listen on (e.g., "0.0.0.0" or "127.0.0.1")
	Port      int    // Port to listen on
	DataDir   string // Data directory for storing sqlite database and media files
	JWTSecret string `yaml:"jwt_secret"` // JWT secret key for signing tokens

	// Database configuration
	DBType     string `yaml:"db_type"`     // Database type: "sqlite", "mysql", or "postgres"
	DBHost     string `yaml:"db_host"`     // Database host
	DBPort     int    `yaml:"db_port"`     // Database port
	DBUser     string `yaml:"db_user"`     // Database user
	DBPassword string `yaml:"db_password"` // Database password
	DBName     string `yaml:"db_name"`     // Database name
	DBSSLMode  string `yaml:"db_ssl_mode"` // PostgreSQL SSL mode (for postgres)

	// OIDC configuration
	OIDCEnabled       bool   `yaml:"oidc_enabled"`        // Enable OIDC login
	OIDCIssuerURL     string `yaml:"oidc_issuer_url"`     // Issuer URL for OIDC discovery
	OIDCClientID      string `yaml:"oidc_client_id"`      // OIDC client ID
	OIDCClientSecret  string `yaml:"oidc_client_secret"`  // OIDC client secret
	OIDCAuthURL       string `yaml:"oidc_auth_url"`       // Authorization endpoint (optional, for manual override)
	OIDCTokenURL      string `yaml:"oidc_token_url"`      // Token endpoint (optional, for manual override)
	OIDCUserInfoURL   string `yaml:"oidc_userinfo_url"`   // UserInfo endpoint (optional, for manual override)
	OIDCScopes        string `yaml:"oidc_scopes"`         // Space-separated scopes (default: "openid profile email")
	OIDCEmailClaim    string `yaml:"oidc_email_claim"`    // Claim name for email (default: "email")
	OIDCNameClaim     string `yaml:"oidc_name_claim"`     // Claim name for username (default: "name")
	OIDCGroupClaim    string `yaml:"oidc_group_claim"`    // Claim name for groups (default: "groups")
	OIDCAllowedGroups string `yaml:"oidc_allowed_groups"` // Comma-separated allowed groups (empty = allow all)
	OIDCAutoRedirect  bool   `yaml:"oidc_auto_redirect"`  // Auto-redirect to OIDC provider (skip login page)
	OIDCVerifyEmail   bool   `yaml:"oidc_verify_email"`   // Require email_verified=true in token

	// Global options
	AllowLocalLogin bool `yaml:"allow_local_login"` // Allow password-based login (default: true)
}

// ParseFlags parses command line flags and returns the server configuration
func ParseFlags() *Config {
	configFile := flag.String("c", "", "Path to configuration file (YAML format)")
	addr := flag.String("addr", "", "Address to listen on")
	port := flag.Int("port", 0, "Port to listen on")
	dataDir := flag.String("data", "", "Data directory for storing sqlite database and media files")

	// Parse command line flags
	flag.Parse()

	// Default configuration with environment variable fallback
	cfg := &Config{
		Addr:      getEnvOrDefault("ADDR", *addr, "0.0.0.0"),
		Port:      getIntEnvOrDefault("PORT", *port, 3001),
		DataDir:   getEnvOrDefault("DATA_DIR", *dataDir, "./data"),
		JWTSecret: getEnvOrDefault("JWT_SECRET", "", ""),
	}

	// If config file is specified, load it (overrides env and defaults, but not command line flags)
	if *configFile != "" {
		if err := loadConfigFile(*configFile, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config file %s: %v\n", *configFile, err)
		} else {
			fmt.Printf("Loaded configuration from %s\n", *configFile)
		}
	}

	// Apply environment variable overrides for database config (only if not set by config file or command line)
	applyEnvOverrides(cfg)

	return cfg
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, value, defaultValue string) string {
	if value != "" {
		return value // command line flag takes precedence
	}
	if env := os.Getenv(key); env != "" {
		return env // environment variable next
	}
	return defaultValue // finally use default
}

// getIntEnvOrDefault gets integer environment variable or returns default value
func getIntEnvOrDefault(key string, value, defaultValue int) int {
	if value != 0 {
		return value // command line flag takes precedence
	}
	if env := os.Getenv(key); env != "" {
		var result int
		if _, err := fmt.Sscanf(env, "%d", &result); err == nil {
			return result // environment variable next
		}
	}
	return defaultValue // finally use default
}

// applyEnvOverrides applies environment variable overrides for database configuration
// Only applies if the config field is empty (not set by config file or command line)
func applyEnvOverrides(cfg *Config) {
	// JWT Secret
	if cfg.JWTSecret == "" {
		if env := os.Getenv("JWT_SECRET"); env != "" {
			cfg.JWTSecret = env
		}
	}

	// Database type
	if cfg.DBType == "" {
		if env := os.Getenv("DB_TYPE"); env != "" {
			cfg.DBType = env
		}
	}

	// Database host
	if cfg.DBHost == "" {
		if env := os.Getenv("DB_HOST"); env != "" {
			cfg.DBHost = env
		}
	}

	// Database port
	if cfg.DBPort == 0 {
		if env := os.Getenv("DB_PORT"); env != "" {
			fmt.Sscanf(env, "%d", &cfg.DBPort)
		}
	}

	// Database user
	if cfg.DBUser == "" {
		if env := os.Getenv("DB_USER"); env != "" {
			cfg.DBUser = env
		}
	}

	// Database password
	if cfg.DBPassword == "" {
		if env := os.Getenv("DB_PASSWORD"); env != "" {
			cfg.DBPassword = env
		}
	}

	// Database name
	if cfg.DBName == "" {
		if env := os.Getenv("DB_NAME"); env != "" {
			cfg.DBName = env
		}
	}

	// PostgreSQL SSL mode
	if cfg.DBSSLMode == "" {
		if env := os.Getenv("DB_SSL_MODE"); env != "" {
			cfg.DBSSLMode = env
		}
	}

	// OIDC configuration
	if !cfg.OIDCEnabled {
		if env := os.Getenv("OIDC_ENABLED"); env != "" {
			if b, err := strconv.ParseBool(env); err == nil {
				cfg.OIDCEnabled = b
			}
		}
	}
	if cfg.OIDCIssuerURL == "" {
		if env := os.Getenv("OIDC_ISSUER_URL"); env != "" {
			cfg.OIDCIssuerURL = env
		}
	}
	if cfg.OIDCClientID == "" {
		if env := os.Getenv("OIDC_CLIENT_ID"); env != "" {
			cfg.OIDCClientID = env
		}
	}
	if cfg.OIDCClientSecret == "" {
		if env := os.Getenv("OIDC_CLIENT_SECRET"); env != "" {
			cfg.OIDCClientSecret = env
		}
	}
	if cfg.OIDCAuthURL == "" {
		if env := os.Getenv("OIDC_AUTH_URL"); env != "" {
			cfg.OIDCAuthURL = env
		}
	}
	if cfg.OIDCTokenURL == "" {
		if env := os.Getenv("OIDC_TOKEN_URL"); env != "" {
			cfg.OIDCTokenURL = env
		}
	}
	if cfg.OIDCUserInfoURL == "" {
		if env := os.Getenv("OIDC_USERINFO_URL"); env != "" {
			cfg.OIDCUserInfoURL = env
		}
	}
	if cfg.OIDCScopes == "" {
		if env := os.Getenv("OIDC_SCOPES"); env != "" {
			cfg.OIDCScopes = env
		}
	}
	if cfg.OIDCEmailClaim == "" {
		if env := os.Getenv("OIDC_EMAIL_CLAIM"); env != "" {
			cfg.OIDCEmailClaim = env
		}
	}
	if cfg.OIDCNameClaim == "" {
		if env := os.Getenv("OIDC_NAME_CLAIM"); env != "" {
			cfg.OIDCNameClaim = env
		}
	}
	if cfg.OIDCGroupClaim == "" {
		if env := os.Getenv("OIDC_GROUP_CLAIM"); env != "" {
			cfg.OIDCGroupClaim = env
		}
	}
	if cfg.OIDCAllowedGroups == "" {
		if env := os.Getenv("OIDC_ALLOWED_GROUPS"); env != "" {
			cfg.OIDCAllowedGroups = env
		}
	}
	if !cfg.OIDCAutoRedirect {
		if env := os.Getenv("OIDC_AUTO_REDIRECT"); env != "" {
			if b, err := strconv.ParseBool(env); err == nil {
				cfg.OIDCAutoRedirect = b
			}
		}
	}
	if !cfg.OIDCVerifyEmail {
		if env := os.Getenv("OIDC_VERIFY_EMAIL"); env != "" {
			if b, err := strconv.ParseBool(env); err == nil {
				cfg.OIDCVerifyEmail = b
			}
		}
	}
	if !cfg.AllowLocalLogin {
		if env := os.Getenv("ALLOW_LOCAL_LOGIN"); env != "" {
			if b, err := strconv.ParseBool(env); err == nil {
				cfg.AllowLocalLogin = b
			}
		}
	}
}

// loadConfigFile loads configuration from a YAML file
// It only fills empty fields, preserving values already set (e.g., from command line)
func loadConfigFile(filename string, cfg *Config) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal YAML into a temporary struct to avoid overwriting existing values
	var temp Config
	if err := yaml.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Only apply values from config file if the field is currently empty (zero value)
	// This preserves command line flag values which have highest priority
	if cfg.Addr == "" || cfg.Addr == "0.0.0.0" {
		if temp.Addr != "" {
			cfg.Addr = temp.Addr
		}
	}
	if cfg.Port == 0 {
		if temp.Port != 0 {
			cfg.Port = temp.Port
		}
	}
	if cfg.DataDir == "" || cfg.DataDir == "./data" {
		if temp.DataDir != "" {
			cfg.DataDir = temp.DataDir
		}
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = temp.JWTSecret
	}
	if cfg.DBType == "" {
		cfg.DBType = temp.DBType
	}
	if cfg.DBHost == "" {
		cfg.DBHost = temp.DBHost
	}
	if cfg.DBPort == 0 {
		cfg.DBPort = temp.DBPort
	}
	if cfg.DBUser == "" {
		cfg.DBUser = temp.DBUser
	}
	if cfg.DBPassword == "" {
		cfg.DBPassword = temp.DBPassword
	}
	if cfg.DBName == "" {
		cfg.DBName = temp.DBName
	}
	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = temp.DBSSLMode
	}

	// OIDC configuration
	if !cfg.OIDCEnabled && temp.OIDCEnabled {
		cfg.OIDCEnabled = temp.OIDCEnabled
	}
	if cfg.OIDCIssuerURL == "" && temp.OIDCIssuerURL != "" {
		cfg.OIDCIssuerURL = temp.OIDCIssuerURL
	}
	if cfg.OIDCClientID == "" && temp.OIDCClientID != "" {
		cfg.OIDCClientID = temp.OIDCClientID
	}
	if cfg.OIDCClientSecret == "" && temp.OIDCClientSecret != "" {
		cfg.OIDCClientSecret = temp.OIDCClientSecret
	}
	if cfg.OIDCAuthURL == "" && temp.OIDCAuthURL != "" {
		cfg.OIDCAuthURL = temp.OIDCAuthURL
	}
	if cfg.OIDCTokenURL == "" && temp.OIDCTokenURL != "" {
		cfg.OIDCTokenURL = temp.OIDCTokenURL
	}
	if cfg.OIDCUserInfoURL == "" && temp.OIDCUserInfoURL != "" {
		cfg.OIDCUserInfoURL = temp.OIDCUserInfoURL
	}
	if cfg.OIDCScopes == "" && temp.OIDCScopes != "" {
		cfg.OIDCScopes = temp.OIDCScopes
	}
	if cfg.OIDCEmailClaim == "" && temp.OIDCEmailClaim != "" {
		cfg.OIDCEmailClaim = temp.OIDCEmailClaim
	}
	if cfg.OIDCNameClaim == "" && temp.OIDCNameClaim != "" {
		cfg.OIDCNameClaim = temp.OIDCNameClaim
	}
	if cfg.OIDCGroupClaim == "" && temp.OIDCGroupClaim != "" {
		cfg.OIDCGroupClaim = temp.OIDCGroupClaim
	}
	if cfg.OIDCAllowedGroups == "" && temp.OIDCAllowedGroups != "" {
		cfg.OIDCAllowedGroups = temp.OIDCAllowedGroups
	}
	if !cfg.OIDCAutoRedirect && temp.OIDCAutoRedirect {
		cfg.OIDCAutoRedirect = temp.OIDCAutoRedirect
	}
	if !cfg.OIDCVerifyEmail && temp.OIDCVerifyEmail {
		cfg.OIDCVerifyEmail = temp.OIDCVerifyEmail
	}
	if !cfg.AllowLocalLogin && temp.AllowLocalLogin {
		cfg.AllowLocalLogin = temp.AllowLocalLogin
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
