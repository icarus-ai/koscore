package utils

import (
	"fmt"
	"io"
	"os"
)

func StreamSize(image io.ReadSeeker) int64 {
	_, _ = image.Seek(0, io.SeekStart)
	size, _ := image.Seek(0, io.SeekEnd)
	_, _ = image.Seek(0, io.SeekStart)
	return size
}

func ResolveFileName(fstream io.ReadSeeker, name string) string {
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
