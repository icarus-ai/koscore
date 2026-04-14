package auth

import (
	"bytes"
	"compress/gzip"
	"errors"
	"sync/atomic"

	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/binary/prefix"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
	"github.com/kernel-ai/koscore/utils/types"
	// "github.com/kernel-ai/koscore/utils/comm"
)

func (m *WLoginSigs) clear() {
	m.A1, m.A1Key = nil, make([]byte, 16)
	m.A2, m.A2Key = nil, make([]byte, 16)
	m.D2, m.D2Key = nil, make([]byte, 16)
	m.RandomKey = make([]byte, 16)
	m.TgtgtKey = nil
	m.PsKey = make(types.MapSS)
	//comm.LOGD("(m *WLoginSigs) clear() ")
}

func (m *State) clear() {
	m.Tlv104, m.Tlv547, m.Tlv174, m.Cookie, m.QrSig, m.KeyExchange = nil, nil, nil, "", nil, nil
}
func (m *BotInfo) clear() { m.Age, m.Gender, m.Name, m.Uin, m.Uid = 0, 0, "", 0, "" }

func NewSession() *Session {
	session := &Session{
		Sig:   &WLoginSigs{},
		State: &State{},
		Info:  &BotInfo{},
	}
	session.Clear()
	return session
}

func (m *Session) Clear() {
	m.Sig.clear()
	m.State.clear()
	m.Info.clear()
	m.Sequence = crypto.RandomU32(5000000, 9900000)
}

func (m *Session) GetAndIncreaseSequence() uint32 { return atomic.AddUint32(&m.Sequence, 1) % 0x8000 }
func (m *Session) GetSequence() uint32            { return atomic.LoadUint32(&m.Sequence) % 0x8000 }

func (m *Session) Marshal() []byte {
	data, _ := proto.Marshal(m)
	hash := crypto.MD5Digest(data)
	return binary.NewBuilder().
		WriteLengthBytes(hash, prefix.Int16|prefix.LengthOnly).
		WriteLengthBytes(compress(data), prefix.Int16|prefix.LengthOnly).
		ToBytes()
}

var ErrDataHashMismatch = errors.New("data hash mismatch")

func UnmarshalSigInfo(data []byte, verify bool) (*Session, error) {
	byt := binary.NewReader(data)
	hash := byt.ReadLengthBytes(prefix.Int16 | prefix.LengthOnly)
	data = uncompress(byt.ReadLengthBytes(prefix.Int16 | prefix.LengthOnly))
	if verify && !bytes.Equal(hash, crypto.MD5Digest(data)) {
		return nil, ErrDataHashMismatch
	}
	return proto.Unmarshal[Session](data)
}

func uncompress(src []byte) []byte {
	b := bytes.NewReader(src)
	r, e := gzip.NewReader(b)
	if e == nil {
		defer r.Close()
		var byt bytes.Buffer
		_, _ = byt.ReadFrom(r)
		return byt.Bytes()
	}
	return nil
}

func compress(src []byte) []byte {
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	_, _ = w.Write(src)
	_ = w.Close()
	return buf.Bytes()
}
