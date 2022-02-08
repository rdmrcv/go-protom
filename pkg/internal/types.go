package internal

import (
	"reflect"
	"time"

	"google.golang.org/protobuf/proto"
)

var (
	TProto = reflect.TypeOf((*proto.Message)(nil)).Elem()
	TTime  = reflect.TypeOf(time.Time{})
)
