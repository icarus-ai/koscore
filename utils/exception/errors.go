package exception

import "errors"

var ErrEmptyRsp = errors.New("empty response data")

var (
	ErrAlreadyOnline = errors.New("already online")
	//ErrNotOnline      = errors.New("not online")
	//ErrMemberNotFound = errors.New("member not found")
	//ErrNotExists      = errors.New("not exists")
)

var (
	ErrSessionExpired       = errors.New("session expired")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrPacketDropped        = errors.New("packet dropped")
	//ErrInvalidPacketType    = errors.New("invalid packet type")
)

var ErrDataHashMismatch = errors.New("data hash mismatch")
