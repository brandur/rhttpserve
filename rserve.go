package main

import (
	"log"

	"github.com/brandur/rserve/cmd"
	_ "github.com/brandur/rserve/cmd/all" // import all commands
	_ "github.com/ncw/rclone/fs/all"      // import all fs
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}
