package utils

import (
	"os"
	"strconv"
)

func GetAddr() string {
	if addr, ok := os.LookupEnv("ADDR"); ok {
		return addr
	}

	return "127.0.0.1"
}

func GetPort() uint64 {
	if port, ok := os.LookupEnv("PORT"); ok {
		if parsed, err := strconv.ParseUint(port, 10, 64); err == nil {
			return parsed
		}
	}

	return 7000
}
