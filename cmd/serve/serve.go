package serve

import (
	"log"
	"net/http"

	"github.com/brandur/rserve/cmd"
	"github.com/brandur/rserve/common"
	"github.com/ncw/rclone/fs"
	"github.com/spf13/cobra"
)

func init() {
	cmd.Root.AddCommand(serveCmd)
}

func getParam(w http.ResponseWriter, r *http.Request, name string) (string, bool) {
	param := r.URL.Query().Get(name)
	if param == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Need parameter: " + name))
		return "", false
	}
	return param, true
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	expiresAt, ok := getParam(w, r, "expires_at")
	if !ok {
		return
	}

	signature, ok := getParam(w, r, "signature")
	if !ok {
		return
	}

	_ = common.Verify(r.URL.Path, expiresAt, signature)

	rclonePath := common.GetRemotePath(r.URL.Path)
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
