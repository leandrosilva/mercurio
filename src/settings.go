package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

// GetCurrentEnv as per MERCURIO_ENV environment variable. When not provided, defaults to development
func GetCurrentEnv() string {
	env := os.Getenv("MERCURIO_ENV")
	if env == "" {
		env = "development"
	}
	return env
}

// GetEnvDir as per MERCURIO_ENV_DIR environment variable. When not provided, defaults to current directory (./)
func GetEnvDir() string {
	envDir := os.Getenv("MERCURIO_ENV_DIR")
	if envDir == "" {
		envDir += "./"
	}
	return envDir
}

// LoadEnvironmentVars from .env files according on the provided MERCURIO_ENV. Defaults to development
func LoadEnvironmentVars() {
	envDir := GetEnvDir()
	env := GetCurrentEnv()

	// 1st: Local overrides of environment-specific settings
	file := envDir + ".env." + env + ".local"
	if godotenv.Load(file) == nil {
		log.Printf("File '%s' is loaded", file)
	}

	if env != "test" {
		// 2nd: Local overrides. This file is loaded for all environments except test
		file = envDir + ".env.local"
		if godotenv.Load(file) == nil {
			log.Printf("File '%s' is loaded", file)
		}
	}

	// 3rd: Shared environment-specific settings. May not .gitignore it
	file = envDir + ".env." + env
	if godotenv.Load(file) == nil {
		log.Printf("File '%s' is loaded", file)
	}

	// The original .env file. It depends whether .gitignore it or not
	file = envDir + ".env"
	if godotenv.Load(file) == nil {
		log.Printf("File '%s' is loaded", file)
	}
}

// GetAuthPrivateKey try and read the provided private key from either MERCURIO_AUTH_PK_TEXT or MERCURIO_AUTH_PK_PATH environment variables
func GetAuthPrivateKey() ([]byte, error) {
	privateKeyText := os.Getenv("MERCURIO_AUTH_PK_TEXT")

	// 1st option is a private key provided right away (hopefully not committed on git)
	if privateKeyText != "" {
		return []byte(privateKeyText), nil
	}

	// 2nd option is a private key provided as a file (also not committed on git, please)
	if privateKeyText == "" {
		privateKeyFilePath := os.Getenv("MERCURIO_AUTH_PK_PATH")

		// If neither are provided, returns an error letting operation guys to know what is expected
		if privateKeyFilePath == "" {
			return nil, errors.New("environment variable MERCURIO_AUTH_PK_TEXT or MERCURIO_AUTH_PK_PATH must be provided")
		}

		privateKey, err := ioutil.ReadFile(privateKeyFilePath)
		if err != nil {
			return nil, err
		}

		return privateKey, nil
	}

	// This is probably unreachable, but compiler can't figure it out
	return nil, errors.New("somehow failed to get a private key")
}

// GetDatabaseConnectionString returns the content of MERCURIO_DB_CONN
func GetDatabaseConnectionString() (string, error) {
	databaseConn := os.Getenv("MERCURIO_DB_CONN")
	if databaseConn == "" {
		return "", errors.New("environment variable MERCURIO_DB_CONN must be provided")
	}
	return databaseConn, nil
}

// GetHTTPServerAddress from the environment variables MERCURIO_HTTP_HOST and MERCURIO_HTTP_PORT
func GetHTTPServerAddress() string {
	httpHost := os.Getenv("MERCURIO_HTTP_HOST")
	if httpHost == "" {
		httpHost = "127.0.0.1"
	}

	httpPort := os.Getenv("MERCURIO_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8000"
	}

	return httpHost + ":" + httpPort
}

// GetCORSOptions builds from the content of MERCURIO_CORS_ALLOWED_ORIGINS, MERCURIO_CORS_ALLOWED_HEADERS,
// and MERCURIO_CORS_ALLOWED_METHODS, with broad defaults when missing
func GetCORSOptions() cors.Options {
	allowedOrigins := os.Getenv("MERCURIO_CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}

	allowedHeaders := os.Getenv("MERCURIO_CORS_ALLOWED_HEADERS")
	if allowedHeaders == "" {
		allowedHeaders = "*"
	}

	allowedMethods := os.Getenv("MERCURIO_CORS_ALLOWED_METHODS")
	if allowedMethods == "" {
		allowedMethods = "GET,POST,PUT,DELETE,HEAD,OPTIONS"
	}

	options := cors.Options{
		AllowedOrigins: strings.Split(allowedOrigins, ","),
		AllowedHeaders: strings.Split(allowedHeaders, ","),
		AllowedMethods: strings.Split(allowedMethods, ","),
	}

	return options
}

// UseMQ as per MERCURIO_MQ missing or equals to "on"
func UseMQ() bool {
	mq := strings.ToLower(os.Getenv("MERCURIO_MQ"))
	if mq == "" || mq == "on" {
		return true
	}
	return false
}

// GetMQSettings builds from the content of MERCURIO_MQ_CONN, MERCURIO_MQ_TOPIC and MERCURIO_MQ_ROUTING_KEY
func GetMQSettings() (MessageQueueSettings, error) {
	mqUse := UseMQ()

	mqConn := os.Getenv("MERCURIO_MQ_CONN")
	if mqConn == "" && mqUse {
		return MessageQueueSettings{}, errors.New("environment variable MERCURIO_MQ_CONN must be provided")
	}

	mqTopic := os.Getenv("MERCURIO_MQ_TOPIC")
	if mqTopic == "" && mqUse {
		return MessageQueueSettings{}, errors.New("environment variable MERCURIO_MQ_TOPIC must be provided")
	}

	mqRoutingKey := os.Getenv("MERCURIO_MQ_ROUTING_KEY")
	if mqRoutingKey == "" && mqUse {
		return MessageQueueSettings{}, errors.New("environment variable MERCURIO_MQ_ROUTING_KEY must be provided")
	}

	settings := MessageQueueSettings{
		Use:        mqUse,
		URL:        mqConn,
		Topic:      mqTopic,
		RoutingKey: mqRoutingKey,
	}

	return settings, nil
}
