package config

import (
	"log"
	"os"
)

// JWTSecret holds the HMAC secret used to sign tokens.
var JWTSecret []byte

// Init loads config values from environment and validates them.
func Init() {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		// 在开发环境允许使用一个弱默认密钥，但会记录警告。
		// 强烈建议在生产环境中通过环境变量注入一个安全的密钥。
		log.Println("WARNING: JWT_SECRET not set — using insecure default for development")
		s = "dev-secret-do-not-use-in-production"
	}
	JWTSecret = []byte(s)
}
