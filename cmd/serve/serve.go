package serve

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
rserve serve starts an HTTP server to serve files from a remote for requests with a valid signature.

Example usage:

	PORT=8090 rserve serve
`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(0, 0, command, args)

		var conf Config
		err := envdecode.Decode(&conf)
		if err != nil {
			common.ExitWithError(err)
		}

		// Check that the remote is configured in env. We don't actually need
		// to read it htough; that will be handled internally by rclone.
		envRemoteName := "RCLONE_CONFIG_" + strings.ToUpper(strings.Replace(conf.Remote+"_TYPE", "-", "_", -1))
		_, found := os.LookupEnv(envRemoteName)
		if !found {
			common.ExitWithError(fmt.Errorf(`variable "%v" is missing`, envRemoteName))
		}

		publicKey, err := base64.URLEncoding.DecodeString(conf.PublicKey)
		if err != nil {
			common.ExitWithError(err)
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

// Config stores the configuration required by the serve command.
type Config struct {
	Port      string `env:"PORT,default=8090"`
	PublicKey string `env:"RSERVE_PUBLIC_KEY,required"`
	Remote    string `env:"RSERVE_REMOTE,required"`
}

// FileServer is a basic encapsulation of the necessary information to serve a
// file out of an rclone remote.
type FileServer struct {
	PublicKey ed25519.PublicKey
	Remote    string
}

// ServeFile serves a file out of an rclone remote based on the request path
// and whether a valid signature was included.
func (s *FileServer) ServeFile(w http.ResponseWriter, r *http.Request) {
	// Don't serve non-GET|HEAD or anything at root (because we know it's not a
	// file).
	if r.Method != "GET" && r.Method != "HEAD" || r.URL.Path == "/" {
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

	// Rclone (or more specifically, newFsSrc, which is copied from rclone)
	// mutates config between runs on single files in a way that doesn't
	// allow it to be run twice in succession, so reset filters between runs.
	fs.Config.Filter, err = fs.NewFilter()
	if err != nil {
		log.Printf("Failed to load filters: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(""))
	}

	fsrc := cmd.NewFsSrc([]string{rclonePath})

	numObjects, size, err := fs.Count(fsrc)

	if err == fs.ErrorDirNotFound {
		if cmd.Verbose {
			log.Printf("No such object")
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No such object"))
		return
	} else if err != nil {
		log.Printf("Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(""))
		return
	}

	if numObjects > 1 {
		if cmd.Verbose {
			log.Printf("Can't serve directory")
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Can only serve single files"))
		return
	}

	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	if cmd.Verbose {
		log.Printf("Set size to %v (%v bytes)",
			fs.SizeSuffix(size).Unit("Bytes"), size)
	}

	if r.Method == "HEAD" {
		log.Printf("Serving HEAD: %s", rclonePath)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
		return
	}

	log.Printf("Serving: %s", rclonePath)
	err = fs.Cat(fsrc, w)
	if err != nil {
		panic(err)
	}

	log.Printf("Successfully served: %s", rclonePath)
}

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
