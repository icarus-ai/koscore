package wtlogin

type EncryptMethod uint8

const (
	EM_ST      EncryptMethod = 0x45
	EM_ECDH    EncryptMethod = 0x07
	EM_ECDH_ST EncryptMethod = 0x87 // same with EM_ECDH, but controlled with a flag, if flag is set to 1, the ST would be used
)
