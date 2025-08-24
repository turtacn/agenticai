// cmd/actl/main.go
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/turtacn/agenticai/cmd/actl/commands"
	"github.com/turtacn/agenticai/internal/logger"
)

const binName = "actl"

var (
	version   = "dev"
	buildDate = "none"
	commitSHA = "none"
)

func main() {
	rootCmd := commands.NewRootCmd(version, buildDate, commitSHA)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatalf("❌  %v", err)
	}
}

// build-time ldflags
func setVersion(flagSet *flag.FlagSet) {
	if flagSet != nil {
		flagSet.String("version", version, "print version and exit")
	}
}

func printVersion() {
	log.Printf("%s version %s (built:%s sha:%s)\n", binName, version, buildDate, commitSHA)
}
//Personal.AI order the ending
