package prefix

type Prefix uint16

const (
	None       Prefix = 0b0000
	Int8       Prefix = 0b0001
	Int16      Prefix = 0b0010
	Int32      Prefix = 0b0100
	LengthOnly Prefix = 0b0000
	WithPrefix Prefix = 0b1000
)
