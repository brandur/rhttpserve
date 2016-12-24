package generate

import (
	"fmt"
	"os"

	"crypto/rand"
	"encoding/base64"
	"github.com/brandur/rserve/cmd"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

var serveCmd = &cobra.Command{
	Use:   "generate",
	Short: `Generates a public/private key pair.`,
	Long: `
rserve generate generates a public/private key pair that can
be used to sign and verify requests to and from the program.
`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(0, 0, command, args)

		public, private, err := generate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Printf("RSERVE_PUBLIC_KEY=%s\n", public)
		fmt.Printf("RSERVE_PRIVATE_KEY=%s\n", private)
	},
}

func init() {
	cmd.Root.AddCommand(serveCmd)
}

func generate() (string, string, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}

	publicEncoded := base64.URLEncoding.EncodeToString(public)
	privateEncoded := base64.URLEncoding.EncodeToString(private)

	return publicEncoded, privateEncoded, nil
}
