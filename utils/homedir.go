package utils

import (
	"fmt"
	"os"
)

func GetHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user's home directory:", err)
		os.Exit(1)
	}
	return homeDir
}
