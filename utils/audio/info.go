package audio

// https://github.com/LagrangeDev/lagrange-python/tree/broken/lagrange/utils/audio

import (
	sysbin "encoding/binary"
	"errors"
	"io"
	"strings"

	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/exception"
)

type Type int

const (
	amr Type = iota
	txSilk
	silkV3
)

type Info struct {
	Type Type
	Time float32
}

func (m Type) String() string {
	switch m {
	case amr:
		return "amr"
	case txSilk, silkV3:
		return "silk"
	default:
		return "unknown"
	}
}

func decode(r io.ReadSeeker, _f bool) (*Info, error) {
	reader := binary.ParseReader(r)
	buf := reader.ReadBytes(1)
	if utils.B2S(buf) != utils.B2S([]byte{0x23}) {
		if !_f {
			return decode(r, true)
		}
		return nil, errors.New("unknown audio type")
	}
	buf = append(buf, reader.ReadBytes(5)...)

	switch {
	case strings.HasPrefix(string(buf), "#!AMR\n"):
		return &Info{
			Type: amr,
			Time: float32(len(reader.ReadAll())) / 1607.0,
		}, nil
	case string(buf) == "#!SILK":
		ver := reader.ReadBytes(3)
		if string(ver) != "_V3" {
			return nil, exception.NewFormat("unsupported silk version: %s", utils.B2S(ver))
		}
		data := reader.ReadAll()
		size := len(data)

		var typ Type
		if _f {
			typ = txSilk // txsilk
		} else {
			typ = silkV3
		}

		blks, pos := 0, 0

		for pos+2 < size {
			size := sysbin.LittleEndian.Uint16(data[pos : pos+2])
			if size == 0xFFFF {
				break
			}
			blks++
			pos += int(size) + 2
		}
		return &Info{Type: typ, Time: float32(blks) * 0.02}, nil
	default:
		return nil, errors.New("unknown audio type")
	}
}

func Decode(r io.ReadSeeker) (*Info, error) {
	defer func() { _, _ = r.Seek(0, io.SeekStart) }()
	_, _ = r.Seek(0, io.SeekStart)
	return decode(r, false)
}
