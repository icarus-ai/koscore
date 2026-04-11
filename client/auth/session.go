package auth

import (
	"os"
	"sync/atomic"

	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/proto"
	"github.com/kernel-ai/koscore/utils/types"
	//"github.com/kernel-ai/koscore/utils/comm"
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

func NewSession() *Session {
	session := &Session{
		Sig:      &WLoginSigs{},
		State:    &State{},
		Info:     &BotInfo{},
		Sequence: crypto.RandomU32(5000000, 9900000),
	}
	session.Sig.clear()
	return session
}

func (m *Session) Clear() {}

func (m *Session) GetAndIncreaseSequence() uint32 { return atomic.AddUint32(&m.Sequence, 1) % 0x8000 }
func (m *Session) GetSequence() uint32            { return atomic.LoadUint32(&m.Sequence) % 0x8000 }

func (m *Session) Save(path string) error {
	d, _ := proto.Marshal(m)
	return os.WriteFile(path, d, 0o644)
}

func LoadSession(path string) (*Session, error) {
	d, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	s, e := proto.Unmarshal[Session](d)
	if e != nil {
		return nil, e
	}
	return s, nil
}
