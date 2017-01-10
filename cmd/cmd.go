package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/ncw/rclone/fs"
	"github.com/spf13/cobra"
)

// Version is rserve's current version number.
const Version = "v0.0.1"

var (
	// Verbose tracks whether a verbose flag was passed into the command line.
	Verbose bool

	// version tracks whether a version flag was passed into the command line.
	version bool
)

// Root is the main rserve command
var Root = &cobra.Command{
	Use:   "rserve",
	Short: "Serves files out of an rclone store.",
	Long: `
Rserve is a command line program to private serve files out of an rclone store.
`,
}

// runRoot implements the main rclone command with no subcommands
func runRoot(cmd *cobra.Command, args []string) {
	if version {
		ShowVersion()
		os.Exit(0)
	} else {
		_ = Root.Usage()
		fmt.Fprintf(os.Stderr, "Command not found.\n")
		os.Exit(1)
	}
}

func init() {
	Root.Run = runRoot
	Root.Flags().BoolVarP(&version, "version", "V", false, "Print the version number")
	Root.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Verbose output")
	cobra.OnInitialize(initConfig)
}

// NewFsSrc creates a new src fs from the arguments
func NewFsSrc(args []string) fs.Fs {
	fsrc := newFsSrc(args[0])
	fs.CalculateModifyWindow(fsrc)
	return fsrc
}

// CheckArgs checks there are enough arguments and prints a message if not
func CheckArgs(MinArgs, MaxArgs int, cmd *cobra.Command, args []string) {
	if len(args) < MinArgs {
		_ = cmd.Usage()
		fmt.Fprintf(os.Stderr, "Command %s needs %d arguments mininum\n", cmd.Name(), MinArgs)
		os.Exit(1)
	} else if len(args) > MaxArgs {
		_ = cmd.Usage()
		fmt.Fprintf(os.Stderr, "Command %s needs %d arguments maximum\n", cmd.Name(), MaxArgs)
		os.Exit(1)
	}
}

// ShowVersion prints the version to stdout
func ShowVersion() {
	fmt.Printf("rclone %s\n", Version)
}

// initConfig is run by cobra after initialising the flags
func initConfig() {
	// Load the rest of the config now we have started the logger
	fs.LoadConfig()
}

// newFsSrc creates a src Fs from a name
//
// This can point to a file
func newFsSrc(remote string) fs.Fs {
	fsInfo, configName, fsPath, err := fs.ParseRemote(remote)
	if err != nil {
		fs.Stats.Error()
		log.Fatalf("Failed to create file system for %q: %v", remote, err)
	}
	f, err := fsInfo.NewFs(configName, fsPath)
	if err == fs.ErrorIsFile {
		if !fs.Config.Filter.InActive() {
			fs.Stats.Error()
			log.Fatalf("Can't limit to single files when using filters: %v", remote)
		}
		// Limit transfers to this file
		err = fs.Config.Filter.AddFile(path.Base(fsPath))
		// Set --no-traverse as only one file
		fs.Config.NoTraverse = true
	}
	if err != nil {
		fs.Stats.Error()
		log.Fatalf("Failed to create file system for %q: %v", remote, err)
	}
	return f
}
