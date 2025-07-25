package cmd

import (
	"io/fs"
	"os"
)

// GetFrontendFS returns the frontend filesystem (for now, using OS filesystem)
// TODO: Replace with embedded filesystem for production builds
func GetFrontendFS() (fs.FS, error) {
	// For development, serve from the build directory
	return os.DirFS("../frontend/build"), nil
}
