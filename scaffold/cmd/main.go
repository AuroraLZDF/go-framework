// Copyright 2022 Innkeeper auroralzdf auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/AuroraLZDF/go-framework/core/version"
	"github.com/AuroraLZDF/go-framework/core/version/verflag"
	"github.com/AuroraLZDF/go-framework/scaffold/internal/generator"
)

func main() {
	// 解析全局 flag：--version 由 verflag 注册，打印版本后直接退出
	pflag.Parse()
	verflag.PrintAndExitIfRequested()

	args := pflag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "version":
		printVersion()
	case "new":
		if len(args) < 2 {
			printUsage()
			os.Exit(1)
		}
		projectName := args[1]
		if err := generator.NewProject(projectName); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating project: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Project '%s' created successfully!\n", projectName)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  cd %s\n", projectName)
		fmt.Printf("  go mod tidy\n")
		fmt.Printf("  go run cmd/app/main.go\n")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// printVersion 打印版本信息（版本号等由 Makefile 通过 -ldflags 注入）.
func printVersion() {
	text, err := version.Get().Text()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get version: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(string(text))
}

func printUsage() {
	fmt.Println("Framework Scaffold - Create new Go Framework projects")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  scaffold new <project-name|path>  Create a new project")
	fmt.Println("  scaffold version                  Print version information")
	fmt.Println("  scaffold --version                Print version information and quit")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  scaffold new my-app")
	fmt.Println("  scaffold new /path/to/api-service")
	fmt.Println("  scaffold version")
}
