// Package all imports all the commands
package all

import (
	// Active commands
	_ "github.com/brandur/rhttpserve/cmd"
	_ "github.com/brandur/rhttpserve/cmd/generate"
	_ "github.com/brandur/rhttpserve/cmd/serve"
	_ "github.com/brandur/rhttpserve/cmd/sign"
	_ "github.com/brandur/rhttpserve/cmd/version"
)
