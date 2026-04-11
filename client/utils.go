package client

import (
	"fmt"
	"io"
	"os"
)

func resolveFileName(fstream io.ReadSeeker, name string) string {
	if name == "" {
		switch o := fstream.(type) {
		case *os.File:
			return o.Name()
		default:
			_, _ = fstream.Seek(0, io.SeekStart)
			bs := make([]byte, 16)
			_, _ = fstream.Read(bs)
			return fmt.Sprintf("%X", bs)
		}
	}
	return name
}
