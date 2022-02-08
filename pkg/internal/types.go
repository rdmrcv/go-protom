package internal

import (
	"reflect"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	TProtoMessage = reflect.TypeOf((*protoreflect.ProtoMessage)(nil)).Elem()
	TProtoMap     = reflect.TypeOf((*protoreflect.Map)(nil)).Elem()
	TProtoList    = reflect.TypeOf((*protoreflect.List)(nil)).Elem()
	TProtoEnum    = reflect.TypeOf((*protoreflect.Enum)(nil)).Elem()
)
