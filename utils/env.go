package utils

import (
	"os"
	"strconv"
	"time"
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

func GetClientCount() uint64 {
	if client, ok := os.LookupEnv("CLIENT_COUNT"); ok {
		if parsed, err := strconv.ParseUint(client, 10, 64); err == nil {
			return parsed
		}
	}

	return 1
}

func GetDelay() time.Duration {
	if delay, ok := os.LookupEnv("DELAY"); ok {
		if parsed, err := time.ParseDuration(delay); err == nil {
			return parsed
		}
	}

	return time.Second
}

func GetConcurrency() uint64 {
	if con, ok := os.LookupEnv("CONCURRENCY"); ok {
		if parsed, err := strconv.ParseUint(con, 10, 64); err == nil {
			return parsed
		}
	}

	return 8
}

func GetWatcherCount() uint {
	if wc, ok := os.LookupEnv("WATCHER_COUNT"); ok {
		if parsed, err := strconv.ParseUint(wc, 10, 64); err == nil {
			return uint(parsed)
		}
	}

	return 8
}
