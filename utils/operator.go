package utils

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strings"
)

func ReadLine() string {
	input := bufio.NewReader(os.Stdin)
	rs, e := input.ReadString('\n')
	if e != nil {
		return ""
	}
	return strings.TrimSpace(rs)
}

func Bool2Int(v bool) int {
	if v {
		return 1
	}
	return 0
}

func CloseIO(r io.Reader) {
	if p, ok := r.(io.Closer); ok {
		_ = p.Close()
	}
}

func NewUUID() string {
	u := make([]byte, 16)
	_, e := io.ReadFull(rand.Reader, u)
	if e != nil {
		return ""
	}
	u[6] = (u[6] & 0x0f) | 0x40 // Set the version to 4 (randomly generated UUID)
	u[8] = (u[8] & 0x3f) | 0x80 // Set the variant to RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func Ternary[T any](condition bool, t_val, f_val T) T {
	if condition {
		return t_val
	}
	return f_val
}

func LazyTernary[T any](condition bool, t_fn, f_fn func() T) T {
	if condition {
		return t_fn()
	}
	return f_fn()
}

func Map[T any, U any](list []T, mapper func(T) U) []U {
	result := make([]U, len(list))
	for i, v := range list {
		result[i] = mapper(v)
	}
	return result
}
