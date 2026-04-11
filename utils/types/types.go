package types

import (
	"encoding/hex"
	"fmt"

	"github.com/kernel-ai/koscore/utils"
)

type Bytes []byte

//func (m Bytes) ToBytes() []byte { return m }

func (m Bytes) ToLowHexStr() string {
	if len(m) == 0 {
		return ""
	} else {
		return fmt.Sprintf("%02x", m)
	}
}
func (m Bytes) ToUpHexStr() string {
	if len(m) == 0 {
		return ""
	} else {
		return fmt.Sprintf("%02X", m)
	}
}
func (m *Bytes) UnmarshalJSON(d []byte) (e error) {
	if size := len(d); size > 2 {
		d, e = hex.DecodeString(utils.B2S(d[1 : size-1])) // 去除收尾引号
		if e == nil && len(d) > 0 {
			*m = d
		}
	}
	return
}

type Tlvs map[uint16][]byte

type MapSS map[string]string

const ERROR_NOT_IMPL = "未实现该功能"

type Strings []string

func (m Strings) String() (s string) {
	for _, v := range m {
		s += " " + v
	}
	if len(s) > 0 {
		s = s[1:]
	}
	return
}
