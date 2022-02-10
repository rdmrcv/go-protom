package internal

import (
	"reflect"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	TTime = reflect.TypeOf(time.Time{})

	TProtoMessage = reflect.TypeOf((*protoreflect.ProtoMessage)(nil)).Elem()
	TProtoMap     = reflect.TypeOf((*protoreflect.Map)(nil)).Elem()
	TProtoList    = reflect.TypeOf((*protoreflect.List)(nil)).Elem()
	TProtoEnum    = reflect.TypeOf((*protoreflect.Enum)(nil)).Elem()

	TProtoDuration = reflect.TypeOf(&durationpb.Duration{})

	TProtoEmpty     = reflect.TypeOf(&emptypb.Empty{})
	TProtoTimestamp = reflect.TypeOf(&timestamppb.Timestamp{})

	TProtoStruct          = reflect.TypeOf(&structpb.Struct{})
	TProtoStructValue     = reflect.TypeOf(&structpb.Value{})
	TProtoStructListValue = reflect.TypeOf(&structpb.ListValue{})

	TProtoWrappedString = reflect.TypeOf(&wrapperspb.StringValue{})
	TProtoWrappedBytes  = reflect.TypeOf(&wrapperspb.BytesValue{})
	TProtoWrappedBool   = reflect.TypeOf(&wrapperspb.BoolValue{})
	TProtoWrappedUInt32 = reflect.TypeOf(&wrapperspb.UInt32Value{})
	TProtoWrappedInt32  = reflect.TypeOf(&wrapperspb.Int32Value{})
	TProtoWrappedUInt64 = reflect.TypeOf(&wrapperspb.UInt64Value{})
	TProtoWrappedInt64  = reflect.TypeOf(&wrapperspb.Int64Value{})
	TProtoWrappedFloat  = reflect.TypeOf(&wrapperspb.FloatValue{})
	TProtoWrappedDouble = reflect.TypeOf(&wrapperspb.DoubleValue{})
)
