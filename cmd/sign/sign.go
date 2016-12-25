package sign

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/brandur/rserve/cmd"
	"github.com/brandur/rserve/common"
	"github.com/joeshaw/envdecode"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

var (
	curl      bool
	skipCheck bool
)

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: `Creates a shareable link.`,
	Long: `
rserve sign creates a shareable link with a valid signature
and expiry. Its parameter should be a path relative to the
remote's root.

Example usage:

	rserve sign my/file.pdf
`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(1, 99999, command, args)

		var conf Config
		err := envdecode.Decode(&conf)
		if err != nil {
			common.ExitWithError(err)
		}

		privateKey, err := base64.URLEncoding.DecodeString(conf.PrivateKey)
		if err != nil {
			common.ExitWithError(err)
		}

		generator := URLGenerator{
			Host:       conf.Host,
			PrivateKey: ed25519.PrivateKey(privateKey),
		}

		for _, arg := range args {
			// Maybe make this configurable at some point.
			expiresAt := time.Now().Add(48 * time.Hour)

			url, err := generator.Generate(arg, expiresAt)
			if err != nil {
				common.ExitWithError(err)
			}

			// Check that the URL that we just generated and the file that it
			// points to is valid by issuing a HEAD request to the server.
			if !skipCheck {
				err = checkURL(url)
				if err != nil {
					common.ExitWithError(err)
				}
			}

			if curl {
				filename := filepath.Base(arg)
				fmt.Printf("curl -o '%s' '%s'\n", filename, url)
			} else {
				fmt.Printf("%s\n", url)
			}
		}
	},
}

type Config struct {
	Host       string `env:"RSERVE_HOST,required"`
	PrivateKey string `env:"RSERVE_PRIVATE_KEY,required"`
}

type URLGenerator struct {
	Host       string
	PrivateKey ed25519.PrivateKey
}

func (s *URLGenerator) Generate(path string, expiresAt time.Time) (string, error) {
	scheme := "https"
	if s.Host == "localhost" || strings.HasPrefix(s.Host, "localhost:") {
		scheme = "http"
	}

	u := url.URL{
		Host:   s.Host,
		Path:   path,
		Scheme: scheme,
	}

	message := common.Message(path, expiresAt.Unix())
	if cmd.Verbose {
		log.Printf("Message: %v", string(message))
	}

	signature := ed25519.Sign(s.PrivateKey, message)

	u.RawQuery = fmt.Sprintf("expires_at=%v&signature=%v",
		expiresAt.Unix(), base64.URLEncoding.EncodeToString(signature))

	return u.String(), nil
}

func init() {
	cmd.Root.AddCommand(signCmd)
	signCmd.Flags().BoolVar(&curl, "curl", false, "Output as cURL command")
	signCmd.Flags().BoolVar(&skipCheck, "skip-check", false,
		"Skip issuing server check of generated URL")
}

func checkURL(url string) error {
	resp, err := http.Head(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Re-request with GET so we can see a response body.
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		message, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf(string(message))
	}

	return nil
}
