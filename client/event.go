package client

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/events.go

import (
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kernel-ai/koscore/client/event"
	"github.com/kernel-ai/koscore/message"
)

var __event_mu sync.RWMutex // protected all EventHandle, since write is very rare, use only one lock to save memory

type EventHandle[T any] struct {
	handlers []func(client *QQClient, event T)
}

func (handle *EventHandle[T]) Subscribe(handler func(client *QQClient, event T)) {
	__event_mu.Lock()
	defer __event_mu.Unlock()
	// shrink the slice
	newHandlers := make([]func(client *QQClient, event T), len(handle.handlers)+1)
	copy(newHandlers, handle.handlers)
	newHandlers[len(handle.handlers)] = handler
	handle.handlers = newHandlers
}

func (handle *EventHandle[T]) dispatch(client *QQClient, event T) {
	__event_mu.RLock()
	defer func() {
		__event_mu.RUnlock()
		if e := recover(); e != nil {
			client.LOGE("event error: %v\n%s", e, debug.Stack())
		}
	}()
	for _, fn := range handle.handlers {
		fn(client, event)
	}
}

type Events struct {
	// event handles
	GroupMessage   EventHandle[*message.GroupMessage]
	PrivateMessage EventHandle[*message.PrivateMessage]
	TempMessage    EventHandle[*message.TempMessage]

	SelfGroupMessage   EventHandle[*message.GroupMessage]
	SelfPrivateMessage EventHandle[*message.PrivateMessage]
	SelfTempMessage    EventHandle[*message.TempMessage]

	GroupJoin  EventHandle[*event.GroupMemberIncrease] // bot进群
	GroupLeave EventHandle[*event.GroupMemberDecrease] // bot 退群

	GroupInvited                 EventHandle[*event.GroupInvite]            // 被邀请入群
	GroupMemberJoinRequest       EventHandle[*event.GroupMemberJoinRequest] // 加群申请
	GroupMemberJoin              EventHandle[*event.GroupMemberIncrease]    // 成员入群
	GroupMemberLeave             EventHandle[*event.GroupMemberDecrease]    // 成员退群
	GroupMute                    EventHandle[*event.GroupMute]
	GroupDigest                  EventHandle[*event.GroupDigestEvent] // 精华消息
	GroupRecall                  EventHandle[*event.GroupRecall]
	GroupMemberPermissionChanged EventHandle[*event.GroupMemberPermissionChanged]
	GroupNameUpdated             EventHandle[*event.GroupNameUpdated]
	GroupReaction                EventHandle[*event.GroupReactionEvent]
	MemberSpecialTitleUpdated    EventHandle[*event.MemberSpecialTitleUpdated]
	NewFriend                    EventHandle[*event.NewFriend]
	NewFriendRequest             EventHandle[*event.NewFriendRequest] // 好友申请
	FriendRecall                 EventHandle[*event.FriendRecall]
	Rename                       EventHandle[*event.Rename]
	FriendNotify                 EventHandle[event.INotifyEvent]
	GroupNotify                  EventHandle[event.INotifyEvent]

	// client event
	Disconnected EventHandle[*event.Disconnected]

	FriendPoke       EventHandle[*event.FriendPokeEvent]
	FriendPokeRecall EventHandle[*event.FriendPokeRecallEvent]
	GroupPoke        EventHandle[*event.GroupPokeEvent]
	GroupPokeRecall  EventHandle[*event.GroupPokeRecallEvent]

	MessageReceived atomic.Int64
	LastMessageTime atomic.Int64
}

func newEventCall() *Events {
	ev := &Events{}
	ev.SelfGroupMessage.Subscribe(func(_ *QQClient, _ *message.GroupMessage) {
		ev.MessageReceived.Add(1)
		ev.LastMessageTime.Store(time.Now().Unix())
	})
	ev.GroupMessage.Subscribe(func(_ *QQClient, _ *message.GroupMessage) {
		ev.MessageReceived.Add(1)
		ev.LastMessageTime.Store(time.Now().Unix())
	})
	ev.SelfPrivateMessage.Subscribe(func(_ *QQClient, _ *message.PrivateMessage) {
		ev.MessageReceived.Add(1)
		ev.LastMessageTime.Store(time.Now().Unix())
	})
	ev.PrivateMessage.Subscribe(func(_ *QQClient, _ *message.PrivateMessage) {
		ev.MessageReceived.Add(1)
		ev.LastMessageTime.Store(time.Now().Unix())
	})
	ev.SelfTempMessage.Subscribe(func(_ *QQClient, _ *message.TempMessage) {
		ev.MessageReceived.Add(1)
		ev.LastMessageTime.Store(time.Now().Unix())
	})
	ev.TempMessage.Subscribe(func(_ *QQClient, _ *message.TempMessage) {
		ev.MessageReceived.Add(1)
		ev.LastMessageTime.Store(time.Now().Unix())
	})
	return ev
}
