package generate

import (
	"fmt"

	"crypto/rand"
	"encoding/base64"
	"github.com/brandur/rhttpserve/cmd"
	"github.com/brandur/rhttpserve/common"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

var serveCmd = &cobra.Command{
	Use:   "generate",
	Short: `Generates a public/private key pair.`,
	Long: `
Generates a public/private key pair that can be used to sign and verify
requests to and from the program.
`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(0, 0, command, args)

		public, private, err := generate()
		if err != nil {
			common.ExitWithError(err)
		}

		fmt.Printf("RHTTPSERVE_PUBLIC_KEY=%s\n", public)
		fmt.Printf("RHTTPSERVE_PRIVATE_KEY=%s\n", private)
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
