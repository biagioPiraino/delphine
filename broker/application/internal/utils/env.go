package utils

import (
	"bufio"
	"os"
	"strings"
)

// strict, failing to load env variables results in panic
func LoadEnvVariables() {
	// TODO: implement tiering for test and production
	filename := ".env"

	file, err := os.OpenFile(filename, os.O_RDONLY, 0440)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())       // get trimmed line
		if line == "" || strings.HasPrefix(line, "#") { // avoid parsing empty lines or comments
			continue
		}

		parts := strings.SplitN(line, "=", 2) // split string on = and gather kv pair
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		os.Setenv(key, value)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
