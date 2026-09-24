package probe

import "os"

// Load returns the contents of path.
func Load(path string) ([]byte, error) {
	return os.ReadFile(path)
}
