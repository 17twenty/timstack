package env

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var FatalOnMissingEnv bool

// LoadFromFile - loads .env file by name
func LoadFromFile(path string) error {
	return godotenv.Overload(path)
}

func GetAsString(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists && FatalOnMissingEnv {
		slog.Error("Missing environment variable", "key", key)
		os.Exit(1)
	}
	return value
}

func GetAsStringElseAlt(key string, alt string) string {
	if FatalOnMissingEnv {
		panic("env.FatalOnMissingEnv is incompatible with using *ElseAlt() functions")
	}
	value, exists := os.LookupEnv(key)
	if !exists {
		return alt
	}
	return value
}

func GetAsInt(key string) int {
	valueStr := GetAsString(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil && FatalOnMissingEnv {
		slog.Error("Environment variable is not an integer", "key", key)
		os.Exit(1)
	}
	return value
}

func GetAsIntElseAlt(key string, alt int) int {
	if FatalOnMissingEnv {
		panic("env.FatalOnMissingEnv is incompatible with using *ElseAlt() functions")
	}
	valueStr := GetAsString(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return alt
	}
	return value
}

func GetAsBool(key string) bool {
	valueStr := GetAsString(key)
	value, err := strconv.ParseBool(valueStr)
	if err != nil && FatalOnMissingEnv {
		slog.Error("Environment variable is not a boolean", "key", key)
		os.Exit(1)
	}
	return value
}
func GetAsBoolElseAlt(key string, alt bool) bool {
	if FatalOnMissingEnv {
		panic("env.FatalOnMissingEnv is incompatible with using *ElseAlt() functions")
	}
	valueStr := GetAsString(key)
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return alt
	}
	return value
}

func GetAsSlice(name string, sep string) []string {
	valStr := GetAsString(name)
	return strings.Split(valStr, sep)
}

func GetAsSliceElseAlt(name string, sep string, alt []string) []string {
	if FatalOnMissingEnv {
		panic("env.FatalOnMissingEnv is incompatible with using *ElseAlt() functions")
	}
	valStr := GetAsString(name)
	if len(valStr) == 0 {
		return alt
	}
	return strings.Split(valStr, sep)
}
