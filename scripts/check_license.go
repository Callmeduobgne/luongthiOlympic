package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	targetExts = map[string]bool{
		".go":  true,
		".ts":  true,
		".tsx": true,
		".js":  true,
		".sh":  true,
	}
	ignoreDirs = map[string]bool{
		".git":         true,
		"node_modules": true,
		"dist":         true,
		"vendor":       true,
		".idea":        true,
		".vscode":      true,
		"coverage":     true,
		"build":        true,
	}
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if ignoreDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if !targetExts[ext] {
			return nil
		}

		// Check for license header
		hasLicense, err := checkLicense(path)
		if err != nil {
			fmt.Printf("Error checking %s: %v\n", path, err)
			return nil
		}

		if !hasLicense {
			fmt.Printf("Missing license: %s\n", path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
	}
}

func checkLicense(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		if lineCount > 20 { // Check first 20 lines
			break
		}
		line := scanner.Text()
		if strings.Contains(line, "Copyright") || strings.Contains(line, "License") {
			return true, nil
		}
	}

	return false, scanner.Err()
}
