package request

import (
	"jurpc/codec"
	"reflect"
)

type Request struct {
	Header *codec.Header
	Argv   reflect.Value
	Replyv reflect.Value
}
