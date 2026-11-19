// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/pflag"

	"github.com/AuroraLZDF/go-framework/core/database/migrate"
	"github.com/AuroraLZDF/go-framework/core/version"
	"github.com/AuroraLZDF/go-framework/scaffold/internal/generator"
)

func main() {
	// 手动分发子命令：各子命令自带 flag（如 --dir/--dsn/--fields），
	// 不能用顶层 pflag.Parse 全局解析，否则未知 flag 会直接报错
	args := os.Args[1:]
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	// --version 打印版本后退出（与 `scaffold version` 等价）
	if args[0] == "--version" {
		printVersion()
		return
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
	case "migrate":
		runMigrate(args[1:])
	case "generate":
		runGenerate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// runMigrate 处理 `scaffold migrate <create|up|down|version>`.
func runMigrate(args []string) {
	if len(args) < 1 {
		migrateUsage()
		os.Exit(1)
	}

	sub := args[0]
	switch sub {
	case "create":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: migration name is required, e.g. scaffold migrate create create_users_table")
			os.Exit(1)
		}
		fs := pflag.NewFlagSet("migrate create", pflag.ExitOnError)
		dir := fs.String("dir", "migrations", "migrations directory")
		_ = fs.Parse(args[2:])

		up, err := migrate.Create(*dir, args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating migration: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Migration created:\n  %s\n  %s\n", up, strings.TrimSuffix(up, ".up.sql")+".down.sql")
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Edit the up/down SQL files")
		fmt.Println("  2. Apply: scaffold migrate up --dsn <DSN>")
	case "up":
		opts := migrateFlags("migrate up", args[1:])
		if err := migrate.Up(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error applying migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Migrations applied")
	case "down":
		fs := pflag.NewFlagSet("migrate down", pflag.ExitOnError)
		dir := fs.String("dir", "migrations", "migrations directory")
		dsn := fs.String("dsn", os.Getenv("FRAMEWORK_MIGRATE_DSN"), "MySQL DSN (or set FRAMEWORK_MIGRATE_DSN)")
		_ = fs.Parse(args[1:])

		steps := 1
		if fs.NArg() > 0 {
			n, err := strconv.Atoi(fs.Arg(0))
			if err != nil || n < 0 {
				fmt.Fprintf(os.Stderr, "Error: invalid steps %q, want a non-negative integer (0 = rollback all)\n", fs.Arg(0))
				os.Exit(1)
			}
			steps = n
		}

		opts := &migrate.Options{Dir: *dir, DSN: *dsn}
		if err := migrate.Down(opts, steps); err != nil {
			fmt.Fprintf(os.Stderr, "Error rolling back migrations: %v\n", err)
			os.Exit(1)
		}
		if steps <= 0 {
			fmt.Println("✅ Rolled back to initial state")
		} else {
			fmt.Printf("✅ Rolled back %d step(s)\n", steps)
		}
	case "version":
		opts := migrateFlags("migrate version", args[1:])
		version, dirty, err := migrate.Version(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading migration version: %v\n", err)
			os.Exit(1)
		}
		if version == 0 {
			fmt.Println("No migrations applied yet")
		} else {
			fmt.Printf("Version: %d (dirty: %v)\n", version, dirty)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown migrate subcommand: %s\n", sub)
		migrateUsage()
		os.Exit(1)
	}
}

// migrateFlags 解析 up/version 共用的 --dir/--dsn 标志.
func migrateFlags(name string, args []string) *migrate.Options {
	fs := pflag.NewFlagSet(name, pflag.ExitOnError)
	dir := fs.String("dir", "migrations", "migrations directory")
	dsn := fs.String("dsn", os.Getenv("FRAMEWORK_MIGRATE_DSN"), "MySQL DSN (or set FRAMEWORK_MIGRATE_DSN)")
	_ = fs.Parse(args)

	return &migrate.Options{Dir: *dir, DSN: *dsn}
}

func migrateUsage() {
	fmt.Println("Database migrations (golang-migrate based)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  scaffold migrate create <name> [--dir migrations]        Create a new up/down migration pair")
	fmt.Println("  scaffold migrate up [--dir migrations] --dsn <DSN>       Apply all pending migrations")
	fmt.Println("  scaffold migrate down [steps] [--dir migrations] --dsn <DSN>")
	fmt.Println("                                                           Roll back N steps (default 1; 0 = all)")
	fmt.Println("  scaffold migrate version [--dir migrations] --dsn <DSN>  Show current migration version")
	fmt.Println()
	fmt.Println("The --dsn flag can be omitted if FRAMEWORK_MIGRATE_DSN is set.")
	fmt.Println("Multi-statement migrations require multiStatements=true in the DSN.")
}

// runGenerate 处理 `scaffold generate <model|crud> <name>`.
func runGenerate(args []string) {
	if len(args) < 2 {
		generateUsage()
		os.Exit(1)
	}

	kind, name := args[0], args[1]

	fs := pflag.NewFlagSet("generate "+kind, pflag.ExitOnError)
	fields := fs.String("fields", "", "comma-separated field list, e.g. name:string,age:int,birth:time")
	force := fs.Bool("force", false, "overwrite existing files")
	_ = fs.Parse(args[2:])

	if *fields == "" {
		fmt.Fprintln(os.Stderr, "Error: --fields is required, e.g. --fields name:string,email:string,age:int")
		os.Exit(1)
	}

	f, err := generator.ParseFields(*fields)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	opts := generator.GenerateOptions{
		Name:   name,
		Fields: f,
		Force:  *force,
	}

	var files []string
	switch kind {
	case "model":
		files, err = generator.GenerateModel(opts)
	case "crud":
		files, err = generator.GenerateCRUD(opts)
	default:
		fmt.Fprintf(os.Stderr, "Unknown generate kind: %s (want model or crud)\n", kind)
		generateUsage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", kind, err)
		os.Exit(1)
	}

	fmt.Printf("✅ %s '%s' generated:\n", strings.ToUpper(kind[:1])+kind[1:], name)
	for _, path := range files {
		fmt.Printf("  %s\n", path)
	}
	if kind == "crud" {
		sn := generator.StructName(name)
		lower := lowerFirst(sn)
		fmt.Println("\nNext steps:")
		fmt.Println("  Wire it up in main.go's RegisterRoutes, e.g.:")
		fmt.Printf("    %sStore, err := store.New%sStore(db)\n", lower, sn)
		fmt.Printf("    if err != nil {\n        log.Fatalf(\"Failed to create store: %%v\", err)\n    }\n")
		fmt.Printf("    %sCtrl := controller.New%sController(%sStore)\n", lower, sn, lower)
		fmt.Printf("    %sCtrl.RegisterRoutes(v1)\n", lower)
	}
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func generateUsage() {
	fmt.Println("Code generation")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  scaffold generate model <name> --fields <field:type,...> [--force]")
	fmt.Println("      Generate a GORM model in internal/model/")
	fmt.Println("  scaffold generate crud <name> --fields <field:type,...> [--force]")
	fmt.Println("      Generate model + store (CRUD repository) + controller with routes")
	fmt.Println()
	fmt.Println("Field types: string, text, bool, int, int8, int16, int32, int64,")
	fmt.Println("             uint, uint8, uint16, uint32, uint64, float, float32, float64, time")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  scaffold generate model user --fields name:string,email:string,age:int")
	fmt.Println("  scaffold generate crud product --fields title:string,price:float,stock:int")
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
	fmt.Println("  scaffold migrate <sub-command>    Database migrations (run `scaffold migrate` for help)")
	fmt.Println("  scaffold generate <kind> <name>   Code generation (run `scaffold generate` for help)")
	fmt.Println("  scaffold version                  Print version information")
	fmt.Println("  scaffold --version                Print version information and quit")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  scaffold new my-app")
	fmt.Println("  scaffold new /path/to/api-service")
	fmt.Println("  scaffold migrate create create_users_table")
	fmt.Println("  scaffold generate crud user --fields name:string,age:int")
	fmt.Println("  scaffold version")
}
