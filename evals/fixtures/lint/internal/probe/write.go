package probe

import "os"

// Save writes data to path.
func Save(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return err
	}
	f.Close()
	return nil
}
