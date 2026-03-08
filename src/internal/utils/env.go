package utils

import (
	"fmt"
	"os"
)

func GetEnvVar(n string, panic_ bool) string {
	val := os.Getenv("MY_VAR")
	if val == "" {
		if panic_ {
			panic(fmt.Sprintf("$s environment variable is required", n))
		}
		return ""
	}
	return val
}
