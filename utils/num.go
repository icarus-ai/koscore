package utils

import (
	"encoding/binary"
	"strconv"
)

func S_NUM[T uint32 | uint64](val string) T {
	v, _ := strconv.Atoi(val)
	return T(v)
}

func B_U32(data []byte) uint32 { return binary.BigEndian.Uint32(data) }
func B_U16(data []byte) uint16 { return binary.BigEndian.Uint16(data) }
