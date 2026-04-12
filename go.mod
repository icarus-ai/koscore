module github.com/kernel-ai/koscore

go 1.26

require (
	github.com/RomiChan/protobuf v0.0.0-20240506080415-f2230fb51d73
	github.com/RomiChan/syncx v0.0.0-20240418144900-b7402ffdebc7
	github.com/fumiama/gofastTEA v0.1.3
	github.com/fumiama/imgsz v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/tidwall/gjson v1.18.0
	golang.org/x/sync v0.20.0
)

require (
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	golang.org/x/image v0.38.0 // indirect
)

replace github.com/RomiChan/protobuf v0.0.0-20240506080415-f2230fb51d73 => github.com/icarus-ai/protobuf v0.0.0-20260411101545-5b3e6a4e1ca7
