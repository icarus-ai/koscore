package ntlogin_type

import "github.com/kernel-ai/koscore/client/packets/structs/sso_type"

// protocol pc
var AttributeEasyLogin = sso_type.NewServiceAttributeD2Empty("trpc.login.ecdh.EcdhService.SsoNTLoginEasyLogin")

// protocol pc
var AttributeUnusualEasyLogin = sso_type.NewServiceAttributeD2Empty("trpc.login.ecdh.EcdhService.SsoNTLoginEasyLoginUnusualDevice")
