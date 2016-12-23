package cat

import (
	"log"
	"net/http"

	"github.com/brandur/rserve"
	"github.com/brandur/rserve/cmd"
	"github.com/ncw/rclone/fs"
	"github.com/spf13/cobra"
)

func init() {
	cmd.Root.AddCommand(serveCmd)
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	rclonePath := rserve.GetRemotePath(r.URL.Path)
	log.Printf("Serving: %s", rclonePath)

	fsrc := cmd.NewFsSrc([]string{rclonePath})

	err := fs.Cat(fsrc, w)
	if err != nil {
		panic(err)
	}
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: `Starts an HTTP server to serve files.`,
	Long: `
rclone cat sends any files to standard output.

You can use it like this to output a single file

    rclone cat remote:path/to/file

Or like this to output any file in dir or subdirectories.

    rclone cat remote:path/to/dir

Or like this to output any .txt files in dir or subdirectories.

    rclone --include "*.txt" cat remote:path/to/dir
`,
	Run: func(command *cobra.Command, args []string) {
		addr := ":8090"

		mux := http.NewServeMux()
		mux.HandleFunc("/", serveFile)

		s := &http.Server{
			Addr:    addr,
			Handler: mux,
		}
		log.Printf("Serving on %s", addr)
		log.Fatal(s.ListenAndServe())
	},
}
