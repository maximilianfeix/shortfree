// Command shortfree finds unregistered 1-4 character domains.
//
//	shortfree check q7.tools x4k.app   live check of single domains
//	shortfree scan                     1-3 chars: DNS sweep + RDAP verification
//	shortfree scan4                    4 chars: DNS sweep of one daily shard
//	shortfree export                   write public data files and README tables
//	shortfree merge <dir>...           combine parallel bootstrap outputs
//	shortfree mcp                      MCP server (stdio) for AI assistants
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
)

const userAgent = "shortfree/1.0 (+https://github.com/maximilianfeix/shortfree)"

// DefaultTLDs are the TLDs scanned when -tlds is not given.
var DefaultTLDs = []string{"com", "net", "de", "ai", "io", "dev", "app", "tools"}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var err error
	switch os.Args[1] {
	case "check":
		err = cmdCheck(ctx, os.Args[2:])
	case "scan":
		err = cmdScan(ctx, os.Args[2:])
	case "scan4":
		err = cmdScan4(ctx, os.Args[2:])
	case "export":
		err = cmdExport(os.Args[2:])
	case "pages":
		err = cmdPages(os.Args[2:])
	case "merge":
		err = cmdMerge(os.Args[2:])
	case "mcp":
		err = cmdMCP(ctx, os.Args[2:])
	case "-v", "--version", "version":
		fmt.Println("shortfree", appVersion())
		return
	case "-h", "--help", "help":
		usage()
		return
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "shortfree:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `shortfree finds unregistered 1-4 character domains.

Usage:
  shortfree check <domain>...   live DNS + RDAP check
  shortfree scan  [flags]       sweep 1-3 character names and verify via RDAP
  shortfree scan4 [flags]       sweep one shard of 4 character names
  shortfree export [flags]      write data/ files and README tables
  shortfree pages [flags]       write static TLD pages, sitemap.xml, robots.txt
  shortfree merge [flags] <dir>...  combine outputs of parallel bootstrap jobs
  shortfree mcp                 MCP server on stdio: live check, search, suggestions
  shortfree version             print the version

Run "shortfree <command> -h" for flags.
`)
}

func parseTLDs(s string) []string {
	if s == "" {
		return DefaultTLDs
	}
	var out []string
	for _, t := range strings.Split(s, ",") {
		if t = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(t)), "."); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func newFlags(name string) *flag.FlagSet {
	return flag.NewFlagSet("shortfree "+name, flag.ExitOnError)
}
