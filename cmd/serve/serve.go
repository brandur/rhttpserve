package serve

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/brandur/rserve/cmd"
	"github.com/brandur/rserve/common"
	"github.com/joeshaw/envdecode"
	"github.com/ncw/rclone/fs"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

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
		cmd.CheckArgs(0, 0, command, args)

		var conf Config
		err := envdecode.Decode(&conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}

		publicKey, err := base64.URLEncoding.DecodeString(conf.PublicKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}

		server := FileServer{
			PublicKey: ed25519.PublicKey(publicKey),
			Remote:    conf.Remote,
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/", server.ServeFile)

		s := &http.Server{
			Addr:    ":" + conf.Port,
			Handler: mux,
		}
		log.Printf("Serving on port %s", conf.Port)
		log.Fatal(s.ListenAndServe())
	},
}

func init() {
	cmd.Root.AddCommand(serveCmd)
}

type Config struct {
	Port      string `env:"PORT,default=8090"`
	PublicKey string `env:"RSERVE_PUBLIC_KEY,required"`
	Remote    string `env:"RSERVE_REMOTE,required"`
}

type FileServer struct {
	PublicKey ed25519.PublicKey
	Remote    string
}

func (s *FileServer) ServeFile(w http.ResponseWriter, r *http.Request) {
	// Don't serve non-GET or anything at root (because we know it's not a
	// file).
	if r.Method != "GET" || r.URL.Path == "/" {
		http.NotFound(w, r)
		return
	}

	expiresAtStr, ok := getParam(w, r, "expires_at")
	if !ok {
		return
	}

	signatureEncoded, ok := getParam(w, r, "signature")
	if !ok {
		return
	}
	signatureStr, err := base64.URLEncoding.DecodeString(signatureEncoded)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Couldn't decode signature"))
		return
	}

	expiresAtInt, err := strconv.ParseInt(expiresAtStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Couldn't parse expires_at"))
		return
	}

	expiresAt := time.Unix(expiresAtInt, 0)
	if expiresAt.Before(time.Now()) {
		if cmd.Verbose {
			log.Printf("Stale expires_at")
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Link is no longer valid because expires_at is in the past"))
		return
	}

	message := common.Message(r.URL.Path, expiresAtInt)
	if cmd.Verbose {
		log.Printf("Message: %v", string(message))
	}

	ok = ed25519.Verify(s.PublicKey, message, []byte(signatureStr))
	if !ok {
		if cmd.Verbose {
			log.Printf("Bad signature")
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Signature verification failed"))
		return
	}

	rclonePath := s.Remote + ":" + r.URL.Path
	log.Printf("Serving: %s", rclonePath)

	fsrc := cmd.NewFsSrc([]string{rclonePath})

	err = fs.Cat(fsrc, w)
	if err != nil {
		panic(err)
	}

	log.Printf("Successfully served: %s", rclonePath)
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
