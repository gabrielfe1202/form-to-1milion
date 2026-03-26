package env

import "os"

func Get(key string, fallback ...string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	if len(fallback) > 0 {
		return fallback[0]
	}

	return ""
}
