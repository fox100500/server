package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type ConnectCfg struct {
	IsSendBox    bool
	URLSandBox   string
	URLReal      string
	URLWebSoc    string
	TokenSandBox string
	TokenReal    string
}

type WebCfg struct {
	Port   string
	PortWS string
}

type Config struct {
	Connect ConnectCfg
	Web     WebCfg
}

// New returns a new Config struct
func New() *Config {

	f, err := os.OpenFile(".env", os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening file .env: %v", err)
	}
	defer f.Close()

	return &Config{
		Connect: ConnectCfg{
			IsSendBox:    getEnvAsBool("IS_SEND_BOX", true),
			URLWebSoc:    getEnv("URL_WEBSOC", ""),
			URLSandBox:   getEnv("URL_SENDBOX", ""),
			URLReal:      getEnv("URL_REAL", ""),
			TokenSandBox: getEnv("TOKEN_SENDBOX", ""),
			TokenReal:    getEnv("TOKEN_REAL", ""),
		},
		Web: WebCfg{
			Port:   getEnv("PORT", ":8080"),
			PortWS: getEnv("PORTWS", ":8081"),
		},
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// Helper to read an environment variable into a bool or return default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

// Helper to read an environment variable into a string slice or return default value
func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)
	return val
}
