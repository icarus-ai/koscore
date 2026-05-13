package cache

import (
	"sync/atomic"

	"github.com/RomiChan/syncx"

	"github.com/kernel-ai/koscore/client/entity"
)

type FaceDetailCache struct {
	detail syncx.Map[string, entity.BotFaceDetail] // map[QSid]Emoji
	size   atomic.Uint32
}

func (m *FaceDetailCache) Set(details []entity.BotFaceDetail) {
	for _, detail := range details {
		if _, ok := m.detail.LoadOrStore(detail.QSid, detail); !ok {
			m.size.Add(1)
		}
	}
}

func (m *FaceDetailCache) Get(qsid string) *entity.BotFaceDetail {
	if v, ok := m.detail.Load(qsid); ok {
		return &v
	}
	return nil
}

func (m *FaceDetailCache) IsEmpty() bool { return m.size.Load() > 0 }
