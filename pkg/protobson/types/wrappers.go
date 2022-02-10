package types

import (
	"fmt"
	"reflect"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type decodeBinaryError struct {
	subtype  byte
	typeName string
}

func (d decodeBinaryError) Error() string {
	return fmt.Sprintf("only binary values with subtype 0x00 or 0x02 can be decoded into %s, but got subtype %v", d.typeName, d.subtype)
}

// ProtoWrappersCodec is the Codec used for google.protobuf.Wrappers type.
type ProtoWrappersCodec struct {
}

var _ bsoncodec.ValueEncoder = &ProtoWrappersCodec{}
var _ bsoncodec.ValueDecoder = &ProtoWrappersCodec{}

// NewProtoWrappersCodec returns a ProtoWrappersCodec that uses p for struct tag parsing.
func NewProtoWrappersCodec() *ProtoWrappersCodec {
	return &ProtoWrappersCodec{}
}

// EncodeValue handles encoding of the `google.protobuf.*Value` message.
func (sc *ProtoWrappersCodec) EncodeValue(r bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	var err error

	switch rv := val.Interface().(type) {
	case *wrapperspb.StringValue:
		val, err = internal.PrepareValueForEncode("ProtoWrappersEncode", val, internal.TProtoWrappedString)
		if err != nil {
			break
		}

		return vw.WriteString(rv.GetValue())
	case *wrapperspb.BytesValue:
		val, err = internal.PrepareValueForEncode("ProtoWrappersEncode", val, internal.TProtoWrappedBytes)
		if err != nil {
			break
		}

		return vw.WriteBinary(rv.GetValue())
	case *wrapperspb.BoolValue:
		val, err = internal.PrepareValueForEncode("ProtoWrappersEncode", val, internal.TProtoWrappedBool)
		if err != nil {
			break
		}

		return vw.WriteBoolean(rv.GetValue())
	case *wrapperspb.UInt32Value:
		val, err = internal.PrepareValueForEncode("ProtoWrappersEncode", val, internal.TProtoWrappedUInt32)
		if err != nil {
			break
		}

		// @TODO: check uint encoding
		return vw.WriteInt32(int32(rv.GetValue()))
	case *wrapperspb.Int32Value:
		val, err = internal.PrepareValueForEncode("ProtoWrappersEncode", val, internal.TProtoWrappedInt32)
		if err != nil {
			break
		}

		return vw.WriteInt32(rv.GetValue())
	case *wrapperspb.UInt64Value:
		val, err = internal.PrepareValueForEncode("ProtoWrappersEncode", val, internal.TProtoWrappedUInt64)
		if err != nil {
			break
		}

		// @TODO: check uint encoding
		return vw.WriteInt64(int64(rv.GetValue()))
	case *wrapperspb.Int64Value:
		val, err = internal.PrepareValueForEncode("ProtoWrappersEncode", val, internal.TProtoWrappedInt64)
		if err != nil {
			break
		}

		return vw.WriteInt64(rv.GetValue())
	case *wrapperspb.FloatValue:
		val, err = internal.PrepareValueForEncode("ProtoWrappersEncode", val, internal.TProtoWrappedFloat)
		if err != nil {
			break
		}

		return vw.WriteDouble(float64(rv.GetValue()))
	case *wrapperspb.DoubleValue:
		val, err = internal.PrepareValueForEncode("ProtoWrappersEncode", val, internal.TProtoWrappedDouble)
		if err != nil {
			break
		}

		return vw.WriteDouble(rv.GetValue())
	default:
		return bsoncodec.ValueEncoderError{Name: "ProtoWrappersEncode", Types: []reflect.Type{
			internal.TProtoWrappedString,
			internal.TProtoWrappedBytes,
			internal.TProtoWrappedBool,
			internal.TProtoWrappedUInt32,
			internal.TProtoWrappedInt32,
			internal.TProtoWrappedUInt64,
			internal.TProtoWrappedInt64,
			internal.TProtoWrappedFloat,
			internal.TProtoWrappedDouble,
		}, Received: val}
	}

	return err
}

// DecodeValue handles decoding of the `google.protobuf.*Value` message.
func (sc *ProtoWrappersCodec) DecodeValue(r bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	var err error

	var rv reflect.Value

	switch val.Type() {
	case internal.TProtoWrappedString:
		val, err = internal.PrepareValueForDecode("ProtoWrappersDecode", val, internal.TProtoWrappedString)
		if err != nil {
			break
		}

		v, err := vr.ReadString()
		if err != nil {
			return err
		}

		rv = reflect.ValueOf(v)
	case internal.TProtoWrappedBytes:
		val, err = internal.PrepareValueForDecode("ProtoWrappersDecode", val, internal.TProtoWrappedBytes)
		if err != nil {
			break
		}

		v, subtype, err := vr.ReadBinary()
		if err != nil {
			return err
		}

		if subtype != bsontype.BinaryGeneric && subtype != bsontype.BinaryBinaryOld {
			return decodeBinaryError{subtype: subtype, typeName: "[]byte"}
		}

		rv = reflect.ValueOf(v)
	case internal.TProtoWrappedBool:
		val, err = internal.PrepareValueForDecode("ProtoWrappersDecode", val, internal.TProtoWrappedBool)
		if err != nil {
			break
		}

		v, err := vr.ReadBoolean()
		if err != nil {
			return err
		}

		rv = reflect.ValueOf(v)
	case internal.TProtoWrappedUInt32:
		val, err = internal.PrepareValueForDecode("ProtoWrappersDecode", val, internal.TProtoWrappedUInt32)
		if err != nil {
			break
		}

		// @TODO: check uint decoding
		v, err := vr.ReadInt32()
		if err != nil {
			return err
		}

		rv = reflect.ValueOf(uint32(v))
	case internal.TProtoWrappedInt32:
		val, err = internal.PrepareValueForDecode("ProtoWrappersDecode", val, internal.TProtoWrappedInt32)
		if err != nil {
			break
		}

		v, err := vr.ReadInt32()
		if err != nil {
			return err
		}

		rv = reflect.ValueOf(v)
	case internal.TProtoWrappedUInt64:
		val, err = internal.PrepareValueForDecode("ProtoWrappersDecode", val, internal.TProtoWrappedUInt64)
		if err != nil {
			break
		}

		// @TODO: check uint decoding
		v, err := vr.ReadInt64()
		if err != nil {
			return err
		}

		rv = reflect.ValueOf(uint64(v))
	case internal.TProtoWrappedInt64:
		val, err = internal.PrepareValueForDecode("ProtoWrappersDecode", val, internal.TProtoWrappedInt64)
		if err != nil {
			break
		}

		v, err := vr.ReadInt64()
		if err != nil {
			return err
		}

		rv = reflect.ValueOf(v)
	case internal.TProtoWrappedFloat:
		val, err = internal.PrepareValueForDecode("ProtoWrappersDecode", val, internal.TProtoWrappedFloat)
		if err != nil {
			break
		}

		v, err := vr.ReadDouble()
		if err != nil {
			return err
		}

		rv = reflect.ValueOf(float32(v))
	case internal.TProtoWrappedDouble:
		val, err = internal.PrepareValueForDecode("ProtoWrappersDecode", val, internal.TProtoWrappedDouble)
		if err != nil {
			break
		}

		v, err := vr.ReadDouble()
		if err != nil {
			return err
		}

		rv = reflect.ValueOf(v)
	default:
		return bsoncodec.ValueEncoderError{Name: "ProtoWrappersEncode", Types: []reflect.Type{
			internal.TProtoWrappedString,
			internal.TProtoWrappedBytes,
			internal.TProtoWrappedBool,
			internal.TProtoWrappedUInt32,
			internal.TProtoWrappedInt32,
			internal.TProtoWrappedUInt64,
			internal.TProtoWrappedInt64,
			internal.TProtoWrappedFloat,
			internal.TProtoWrappedDouble,
		}, Received: val}
	}

	val.Elem().FieldByName("Value").Set(rv)

	return err
}
