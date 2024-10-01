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
	DBDsn                     string             `env:"DATABASE_URI"`
	Addr                      netaddr.NetAddress `env:"RUN_ADDRESS"`
	AccuralSystemAddr         netaddr.NetAddress `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel                  string             `env:"LOG_LEVEL"`
	JWTSecretKey              string             `env:"SECRET_KEY"`
	TokenExp                  time.Duration      `env:"TOKEN_EXP"`
	ConnectionsLimit          int                `env:"CONNECTIONS_LIMIT"`
	OrderUpdaterRateLimit     int                `env:"ORDER_UPDATER_RATE_LIMIT"`
	OrderGetterOrdersLimit    int                `env:"ORDER_GETTER_ORDERS_LIMIT"`
	OrderGetterRateLimit      int                `env:"ORDER_GETTER_RATE_LIMIT"`
	OrderSenderRateLimit      int                `env:"ORDER_SENDER_RATE_LIMIT"`
	OrderSenderAccrualTimeout time.Duration      `env:"ORDER_SENDER_ACCRUAL_TIMEOUT"`
	OrderSenderAccrualRetries int                `env:"ORDER_SENDER_ACCRUAL_RETRIES"`
	OrderEnricherTimeout      time.Duration      `env:"ORDER_ENRICHER_TIMEOUT"`
	OrderEnricherPeriod       time.Duration      `env:"ORDER_ENRICHER_PERIOD"`
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
	flag.StringVar(&cfg.DBDsn, "d", "", "database connection string")
	flag.StringVar(&cfg.LogLevel, "log-level", "INFO", "Log level")
	flag.StringVar(&cfg.JWTSecretKey, "secret-key", generateJWTKey(), "jwt secret key")
	flag.DurationVar(&cfg.TokenExp, "token-exp", time.Hour*24, "token expiration time")
	flag.IntVar(&cfg.ConnectionsLimit, "connections-limit", 100, "limit of postgres connections")
	flag.DurationVar(&cfg.OrderEnricherPeriod, "order-enricher-period", time.Second*10, "period of running order enricher")
	flag.DurationVar(&cfg.OrderEnricherTimeout, "order-enricher-timeout", time.Second*10, "timeout for one iteration in order enricher")
	flag.IntVar(&cfg.OrderSenderAccrualRetries, "order-sender-accrual-retries", 3, "retries num for send orders to accrual service in order sender")
	flag.IntVar(&cfg.OrderSenderRateLimit, "order-sender-rate-limit", 10, "rate limit for send orders to accrual service in order sender")
	flag.DurationVar(&cfg.OrderSenderAccrualTimeout, "order-sender-accrual-timeout", time.Millisecond*500, "timeout for single request in order sender")
	flag.IntVar(&cfg.OrderUpdaterRateLimit, "order-updater-rate-limit", 10, "rate limit for updating db in order updater")
	flag.IntVar(&cfg.OrderGetterOrdersLimit, "order-getter-orders-limit", 1000, "num of orders in one iteration in order getter")
	flag.IntVar(&cfg.OrderGetterRateLimit, "order-getter-rate-limit", 10, "rate limit for getting orders in order getter")
	flag.Parse()
}

func parseEnv(cfg *Config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatalf("Can't parse env variables: %s", err)
	}
}
