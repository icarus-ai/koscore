package event

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/entities.go

import "errors"

var (
	ErrAlreadyOnline  = errors.New("already online")
	ErrNotOnline      = errors.New("not online")
	ErrMemberNotFound = errors.New("member not found")
	ErrNotExists      = errors.New("not exists")
)

var (
	ErrSessionExpired       = errors.New("session expired")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrPacketDropped        = errors.New("packet dropped")
	ErrInvalidPacketType    = errors.New("invalid packet type")
)

type Disconnected struct {
	Message string
}
