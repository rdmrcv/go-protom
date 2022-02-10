package types

import (
	"errors"
	"reflect"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	ErrUnknownValueKind = errors.New("unknown value kind")
)

// ProtoStructCodec is the Codec used for google.protobuf.Struct type.
type ProtoStructCodec struct {
}

var _ bsoncodec.ValueEncoder = &ProtoStructCodec{}
var _ bsoncodec.ValueDecoder = &ProtoStructCodec{}

// NewProtoStructCodec returns a ProtoStructCodec that uses p for struct tag parsing.
func NewProtoStructCodec() *ProtoStructCodec {
	return &ProtoStructCodec{}
}

// EncodeValue handles encoding of the `google.protobuf.Struct` message.
func (sc *ProtoStructCodec) EncodeValue(r bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	val, err := internal.PrepareValueForEncode("ProtoStructEncode", val, internal.TProtoStruct)
	if err != nil {
		return err
	}

	protoMsg := val.Interface().(*structpb.Struct)

	dw, err := vw.WriteDocument()
	if err != nil {
		return err
	}

	encoder, err := r.LookupEncoder(internal.TProtoStructValue)
	if err != nil {
		return err
	}

	var rv reflect.Value

	for key, value := range protoMsg.GetFields() {
		vw, err = dw.WriteDocumentElement(key)
		rv = reflect.ValueOf(value)

		if err := encoder.EncodeValue(bsoncodec.EncodeContext{Registry: r.Registry}, vw, rv); err != nil {
			return err
		}
	}

	return dw.WriteDocumentEnd()
}

// DecodeValue handles decoding of the `google.protobuf.Struct` message.
func (sc *ProtoStructCodec) DecodeValue(r bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	val, err := internal.PrepareValueForDecode("ProtoStructDecodeValue", val, internal.TProtoStruct)
	if err != nil {
		return err
	}

	protoMsg := val.Interface().(*structpb.Struct)

	protoMsg.Reset()

	protoMsg.Fields = map[string]*structpb.Value{}

	decoder, err := r.LookupDecoder(internal.TProtoStructValue)
	if err != nil {
		return err
	}

	dr, err := vr.ReadDocument()
	if err != nil {
		return err
	}

	for {
		name, vr2, err := dr.ReadElement()
		if errors.Is(err, bsonrw.ErrEOD) {
			break
		}
		if err != nil {
			return err
		}

		vc := &structpb.Value{}

		if err := decoder.DecodeValue(
			bsoncodec.DecodeContext{Registry: r.Registry, Ancestor: val.Type()},
			vr2,
			reflect.ValueOf(vc),
		); err != nil {
			return err
		}

		protoMsg.Fields[name] = vc
	}

	return nil
}

// ProtoStructValueCodec is the Codec used for google.protobuf.Value type.
type ProtoStructValueCodec struct {
}

// NewProtoStructValueCodec returns a ProtoStructValueCodec that uses p for struct tag parsing.
func NewProtoStructValueCodec() *ProtoStructValueCodec {
	return &ProtoStructValueCodec{}
}

// EncodeValue handles encoding of the `google.protobuf.Value` message.
func (sc *ProtoStructValueCodec) EncodeValue(r bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	val, err := internal.PrepareValueForEncode("ProtoStructEncode", val, internal.TProtoStructValue)
	if err != nil {
		return err
	}

	protoMsg := val.Interface().(*structpb.Value)

	switch v := protoMsg.GetKind().(type) {
	case *structpb.Value_NullValue:
		return vw.WriteNull()
	case *structpb.Value_NumberValue:
		return vw.WriteDouble(v.NumberValue)
	case *structpb.Value_StringValue:
		return vw.WriteString(v.StringValue)
	case *structpb.Value_BoolValue:
		return vw.WriteBoolean(v.BoolValue)
	case *structpb.Value_StructValue, *structpb.Value_ListValue:
		rv := reflect.ValueOf(v).Elem()
		decoder, err := r.Registry.LookupEncoder(rv.Type())
		if err != nil {
			return err
		}

		return decoder.EncodeValue(bsoncodec.EncodeContext{Registry: r.Registry}, vw, rv)
	default:
		return ErrUnknownValueKind
	}
}

// DecodeValue handles decoding of the `google.protobuf.Value` message.
func (sc *ProtoStructValueCodec) DecodeValue(r bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	val, err := internal.PrepareValueForDecode("ProtoStructDecodeValue", val, internal.TProtoStructValue)
	if err != nil {
		return err
	}

	protoMsg := val.Interface().(*structpb.Value)

	switch vr.Type() {
	case bsontype.Null:
		protoMsg.Kind = &structpb.Value_NullValue{NullValue: structpb.NullValue_NULL_VALUE}
		return vr.ReadNull()
	case bsontype.Int32:
		v, err := vr.ReadInt32()
		if err != nil {
			return err
		}

		protoMsg.Kind = &structpb.Value_NumberValue{NumberValue: float64(v)}
	case bsontype.Int64:
		v, err := vr.ReadInt64()
		if err != nil {
			return err
		}

		protoMsg.Kind = &structpb.Value_NumberValue{NumberValue: float64(v)}
	case bsontype.Double:
		v, err := vr.ReadDouble()
		if err != nil {
			return err
		}

		protoMsg.Kind = &structpb.Value_NumberValue{NumberValue: v}
	case bsontype.String:
		v, err := vr.ReadString()
		if err != nil {
			return err
		}

		protoMsg.Kind = &structpb.Value_StringValue{StringValue: v}
	case bsontype.Boolean:
		v, err := vr.ReadBoolean()
		if err != nil {
			return err
		}

		protoMsg.Kind = &structpb.Value_BoolValue{BoolValue: v}
	case bsontype.Array:
		lv := &structpb.ListValue{}
		rv := reflect.ValueOf(lv)

		protoMsg.Kind = &structpb.Value_ListValue{ListValue: lv}

		decoder, err := r.Registry.LookupDecoder(rv.Type())
		if err != nil {
			return err
		}

		return decoder.DecodeValue(bsoncodec.DecodeContext{Registry: r.Registry, Ancestor: val.Type()}, vr, rv)
	case bsontype.EmbeddedDocument:
		lv := &structpb.Struct{}
		rv := reflect.ValueOf(lv)

		protoMsg.Kind = &structpb.Value_StructValue{StructValue: lv}

		decoder, err := r.Registry.LookupDecoder(rv.Type())
		if err != nil {
			return err
		}

		return decoder.DecodeValue(bsoncodec.DecodeContext{Registry: r.Registry, Ancestor: val.Type()}, vr, rv)
	default:
		return ErrUnknownValueKind
	}

	return nil
}

// ProtoStructListValueCodec is the Codec used for google.protobuf.ListValue type.
type ProtoStructListValueCodec struct {
}

var _ bsoncodec.ValueEncoder = &ProtoStructListValueCodec{}
var _ bsoncodec.ValueDecoder = &ProtoStructListValueCodec{}

// NewProtoStructListValueCodec returns a ProtoStructListValueCodec that uses p for struct tag parsing.
func NewProtoStructListValueCodec() *ProtoStructListValueCodec {
	return &ProtoStructListValueCodec{}
}

// EncodeValue handles encoding of the `google.protobuf.ListValue` message.
func (sc *ProtoStructListValueCodec) EncodeValue(r bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	val, err := internal.PrepareValueForEncode("ProtoStructEncode", val, internal.TProtoStruct)
	if err != nil {
		return err
	}

	protoMsg := val.Interface().(*structpb.ListValue)

	aw, err := vw.WriteArray()
	if err != nil {
		return err
	}

	encoder, err := r.LookupEncoder(internal.TProtoStructValue)
	if err != nil {
		return err
	}

	var rv reflect.Value

	for _, value := range protoMsg.GetValues() {
		vw, err := aw.WriteArrayElement()
		if err != nil {
			return err
		}

		rv = reflect.ValueOf(value)

		if err := encoder.EncodeValue(bsoncodec.EncodeContext{Registry: r.Registry}, vw, rv); err != nil {
			return err
		}
	}

	return aw.WriteArrayEnd()
}

// DecodeValue handles decoding of the `google.protobuf.Struct` message.
func (sc *ProtoStructListValueCodec) DecodeValue(r bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	val, err := internal.PrepareValueForDecode("ProtoStructDecodeValue", val, internal.TProtoStruct)
	if err != nil {
		return err
	}

	protoMsg := val.Interface().(*structpb.ListValue)

	protoMsg.Reset()

	decoder, err := r.LookupDecoder(internal.TProtoStructValue)
	if err != nil {
		return err
	}

	ar, err := vr.ReadArray()
	if err != nil {
		return err
	}

	for {
		vr2, err := ar.ReadValue()
		if errors.Is(err, bsonrw.ErrEOA) {
			break
		}
		if err != nil {
			return err
		}

		vc := &structpb.Value{}

		if err := decoder.DecodeValue(
			bsoncodec.DecodeContext{Registry: r.Registry, Ancestor: val.Type()},
			vr2,
			reflect.ValueOf(vc),
		); err != nil {
			return err
		}

		protoMsg.Values = append(protoMsg.Values, vc)
	}

	return nil
}
