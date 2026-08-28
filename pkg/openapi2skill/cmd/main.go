package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/UnicomAI/wanwu/pkg/openapi2skill"
)

var (
	buildTime    string
	buildVersion string
	gitCommitID  string
	gitBranch    string
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	// Handle --help / -h early
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			printUsage()
			return
		}
		if arg == "-v" || arg == "--version" {
			printVersion()
			return
		}
	}

	var (
		outputDir = "."
		skillName string
		force     bool
		specFile  string
	)

	// Manual argument parsing to support flags after positional arg
	// (e.g. "openapi2skill spec.json -o ./out -n name -f")
	// Go's flag package stops parsing at the first non-flag argument,
	// so we parse manually to match the npm CLI's invocation style.
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "error: -o requires a value")
				os.Exit(1)
			}
			outputDir = args[i]
		case "-n":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "error: -n requires a value")
				os.Exit(1)
			}
			skillName = args[i]
		case "-f":
			force = true
		case "-v", "--version":
			printVersion()
			return
		case "-h", "--help":
			printUsage()
			return
		default:
			if specFile != "" {
				fmt.Fprintf(os.Stderr, "error: unexpected argument %q (spec file already set to %q)\n", args[i], specFile)
				os.Exit(1)
			}
			specFile = args[i]
		}
	}

	if specFile == "" {
		fmt.Fprintln(os.Stderr, "error: spec file path is required")
		printUsage()
		os.Exit(1)
	}

	// Read spec file
	specData, err := os.ReadFile(specFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to read spec file: %v\n", err)
		os.Exit(1)
	}

	// Build options
	opts := openapi2skill.ConvertOptions{
		OutputDir: outputDir,
		Parser: openapi2skill.ParserOptions{
			SkillName: skillName,
			GroupBy:   openapi2skill.GroupByAuto,
		},
	}

	// The -f flag is accepted for backward compatibility with the npm CLI.
	// The Go implementation always overwrites (os.WriteFile truncates),
	// so -f is effectively a no-op.
	_ = force

	// Convert
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := openapi2skill.Convert(ctx, specData, opts); err != nil {
		fmt.Fprintf(os.Stderr, "conversion failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("done")
}

func printVersion() {
	fmt.Printf("openapi2skill %s\n", buildVersion)
	fmt.Printf("  build_time:    %s\n", buildTime)
	fmt.Printf("  git_commit_id: %s\n", gitCommitID)
	fmt.Printf("  git_branch:    %s\n", gitBranch)
}

func printUsage() {
	fmt.Println("openapi2skill - convert OpenAPI 3.x spec to Agent Skills format")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  openapi2skill <spec-file> [-o <dir>] [-n <name>] [-f]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <spec-file>    Path to OpenAPI 3.x specification file (JSON or YAML)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -o <dir>       Output directory (default: .)")
	fmt.Println("  -n <name>      Skill name override (default: auto-detected from info.title)")
	fmt.Println("  -f             Force overwrite existing output")
	fmt.Println("  -v, --version  Print version")
	fmt.Println("  -h, --help     Print help")
}
