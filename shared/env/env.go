package env

import (
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func GetInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	intVal , _ := strconv.ParseInt(value ,10 , 64)

	return int(intVal)
}

func GetBool(key string , fallback bool) bool {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	boolVal , _ := strconv.ParseBool(value)

	return boolVal
}