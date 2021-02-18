package main

import (
	"errors"
	"io/ioutil"
	"os"
)

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
			return nil, errors.New("Environment variable MERCURIO_AUTH_PK_TEXT or MERCURIO_AUTH_PK_PATH must be provided")
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
		return "", errors.New("Environment variable MERCURIO_DB_CONN must be provided")
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
