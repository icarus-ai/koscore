package event

type INotifyEvent interface {
	From() uint64
	Content() string
}
