package main

import (
	"log"

	"github.com/brandur/rhttpserve/cmd"
	_ "github.com/brandur/rhttpserve/cmd/all" // import all commands
	_ "github.com/ncw/rclone/fs/all"          // import all fs
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}
