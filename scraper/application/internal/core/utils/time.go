package utils

import "time"

func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func Today() string {
	return time.Now().UTC().Format("2006-01-02")
}
