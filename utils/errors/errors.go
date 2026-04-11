package errors

type ArgumenttOfRangeException struct {
	message string
}

func (m ArgumenttOfRangeException) Error() string { return m.message }

func NewArgumenttOfRangeException(err string) ArgumenttOfRangeException {
	return ArgumenttOfRangeException{message: err}
}
