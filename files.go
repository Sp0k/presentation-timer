package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func SavePresentationData(data string) string {
	// Get directory path
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = filepath.Join(os.Getenv("HOME"), "go")
	}
	directory := filepath.Join(goPath, "PresentationData")
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		panic(err)
	}

	// Generate file
	now := time.Now()
	filename := fmt.Sprintf("presentation-%s.txt", now.Format("06-01-02-15-04"))
	fullPath := filepath.Join(directory, filename)
	if err := os.WriteFile(fullPath, []byte(data), 0644); err != nil {
		panic(err)
	}

	return directory
}
