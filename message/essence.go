package message

type GroupEssenceMessage struct {
	OperatorUin  uint64
	OperatorUid  string
	OperatorTime uint64
	CanRemove    bool
	Message      *GroupMessage
}
