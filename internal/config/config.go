package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/ry461ch/loyalty_system/internal/models/netaddr"
)

type Config struct {
	DBDsn             string             `env:"DATABASE_URI"`
	Addr              netaddr.NetAddress `env:"RUN_ADDRESS"`
	AccuralSystemAddr netaddr.NetAddress `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel          string             `env:"LOG_LEVEL"`
	JWTSecretKey      string             `env:"SECRET_KEY"`
	TokenExp          time.Duration      `env:"TOKEN_EXP"`
}

func generateJWTKey() string {
	defaultSecretKey := make([]byte, 16)
	rand.Read(defaultSecretKey)
	return hex.EncodeToString(defaultSecretKey)
}

func New() *Config {
	addr := netaddr.NetAddress{Host: "localhost", Port: 8080}
	defaultSecretKey := make([]byte, 16)
	rand.Read(defaultSecretKey)
	cfg := &Config{
		LogLevel:     "INFO",
		Addr:         addr,
		JWTSecretKey: generateJWTKey(),
		TokenExp:     time.Hour * 24,
	}
	parseArgs(cfg)
	parseEnv(cfg)
	return cfg
}

func parseArgs(cfg *Config) {
	flag.Var(&cfg.Addr, "a", "Net address host:port")
	flag.Var(&cfg.AccuralSystemAddr, "r", "Net address of AccuralSystemService host:port")
	flag.StringVar(&cfg.LogLevel, "l", "INFO", "Log level")
	flag.StringVar(&cfg.DBDsn, "d", "", "database connection string")
	flag.StringVar(&cfg.JWTSecretKey, "k", generateJWTKey(), "jwt secret key")
	flag.DurationVar(&cfg.TokenExp, "e", time.Hour*24, "token expiration time")
	flag.Parse()
}

func parseEnv(cfg *Config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatalf("Can't parse env variables: %s", err)
	}
}
