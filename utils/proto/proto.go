package proto

import "github.com/RomiChan/protobuf/proto"

func Unmarshal[T any](data []byte) (*T, error) {
	p := new(T)
	if e := proto.Unmarshal(data, p); e != nil {
		return nil, e
	}
	return any(p).(*T), nil
}

var Marshal = proto.Marshal

var TRUE = proto.Some(true)

func Some[T any](val T) proto.Option[T] { return proto.Some[T](val) }
