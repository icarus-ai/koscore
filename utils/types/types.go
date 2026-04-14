package types

import (
	"encoding/hex"
	"fmt"

	"github.com/kernel-ai/koscore/utils"
	"github.com/pkg/errors"
)

type Bytes []byte

func (m *Bytes) ToBytes() []byte { return *m }
func (m *Bytes) ToLowHexStr() string {
	if len(*m) == 0 {
		return ""
	} else {
		return fmt.Sprintf("%02x", *m)
	}
}
func (m *Bytes) ToUpHexStr() string {
	if len(*m) == 0 {
		return ""
	} else {
		return fmt.Sprintf("%02X", *m)
	}
}
func (m *Bytes) UnmarshalJSON(d []byte) (e error) {
	if size := len(d); size > 2 {
		// 去除收尾引号
		if d, e = hex.DecodeString(utils.B2S(d[1 : size-1])); e != nil {
			return errors.Wrap(e, "unmarshal json")
		}
		if len(d) > 0 {
			*m = d
		}
	}
	return
}
func (m *Bytes) MarshalJSON() ([]byte, error) {
	if len(*m) == 0 {
		return []byte{'"', '"'}, nil
	}
	return utils.S2B(fmt.Sprintf(`"%X"`, *m)), nil // 添加收尾引号
}

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

type Tlvs map[uint16][]byte
type MapSS map[string]string

const ERROR_NOT_IMPL = "未实现该功能"
