package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerAddress      string
	ServerWriteTimeout time.Duration
	ServerReadTimeout  time.Duration
	ServerIdleTimeout  time.Duration
	PgHost             string
	PgPort             string
	PgDb               string
	PgUser             string
	PgPasswd           string
	PgPoolMaxConn      int
	SpaceXUrl          string
}

func (o *Config) DSN() string {
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s pool_max_conns=%d",
		o.PgHost, o.PgPort, o.PgDb, o.PgUser, o.PgPasswd, o.PgPoolMaxConn)
	return dsn
}

func NewConfig() *Config {
	var (
		err                error
		maxConns           int
		serverWriteTimeout time.Duration
		serverReadTimeout  time.Duration
		serverIdleTimeout  time.Duration
	)
	maxConns, err = strconv.Atoi(getEnvOrDefault("MAX_CONNS", "99"))
	if err != nil {
		panic(err)
	}

	serverWriteTimeout, err = getDurationFromEnv("SERVER_WRITE_TIMEOUT", "15s")
	if err != nil {
		panic(err)
	}
	serverReadTimeout, err = getDurationFromEnv("SERVER_READ_TIMEOUT", "15s")
	if err != nil {
		panic(err)
	}
	serverIdleTimeout, err = getDurationFromEnv("SERVER_IDLE_TIMEOUT", "30s")
	if err != nil {
		panic(err)
	}

	cfg := Config{
		ServerAddress:      getEnvOrDefault("SERVER_ADDRESS", ":5000"),
		ServerWriteTimeout: serverWriteTimeout,
		ServerReadTimeout:  serverReadTimeout,
		ServerIdleTimeout:  serverIdleTimeout,
		PgHost:             getEnvOrDefault("POSTGRES_HOST", "localhost"),
		PgPort:             getEnvOrDefault("POSTGRES_PORT", "5432"),
		PgDb:               getEnvOrDefault("POSTGRES_DB", "space"),
		PgUser:             getEnvOrDefault("POSTGRES_USER", "postgres"),
		PgPasswd:           getEnvOrDefault("POSTGRES_PASSWORD", ""),
		PgPoolMaxConn:      maxConns,
		SpaceXUrl:          getEnvOrDefault("SPACEX_URL", "https://api.spacexdata.com/v4"),
	}
	return &cfg
}

func getEnvOrDefault(name, value string) string {
	v := os.Getenv(name)
	if v == "" {
		v = value
	}
	return v
}

func getDurationFromEnv(key, value string) (time.Duration, error) {
	v := getEnvOrDefault(key, value)
	return time.ParseDuration(v)
}
