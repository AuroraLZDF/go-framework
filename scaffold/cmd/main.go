// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/AuroraLZDF/go-framework/scaffold/internal/generator"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	projectName := os.Args[2]

	switch command {
	case "new":
		if err := generator.NewProject(projectName); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating project: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Project '%s' created successfully!\n", projectName)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  cd %s\n", projectName)
		fmt.Printf("  go mod tidy\n")
		fmt.Printf("  cp config.example.yaml config.yaml\n")
		fmt.Printf("  # Edit config.yaml with your settings\n")
		fmt.Printf("  go run cmd/app/main.go\n")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Framework Scaffold - Create new Go Framework projects")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  scaffold new <project-name>  Create a new project")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  scaffold new my-app")
	fmt.Println("  scaffold new api-service")
}
