package binary

import "encoding/binary"

func B2U32(data []byte) uint32 { return binary.BigEndian.Uint32(data) }
func B2U16(data []byte) uint16 { return binary.BigEndian.Uint16(data) }
