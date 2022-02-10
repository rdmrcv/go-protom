package types

import (
	"reflect"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoTimestampCodec is the Codec used for google.protobuf.Timestamp type.
type ProtoTimestampCodec struct {
}

var _ bsoncodec.ValueEncoder = &ProtoTimestampCodec{}
var _ bsoncodec.ValueDecoder = &ProtoTimestampCodec{}

// NewProtoTimestampCodec returns a ProtoEmptyCodec that uses p for struct tag parsing.
func NewProtoTimestampCodec() *ProtoTimestampCodec {
	return &ProtoTimestampCodec{}
}

// EncodeValue handles encoding of the `google.protobuf.Empty` message.
func (sc *ProtoTimestampCodec) EncodeValue(r bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	val, err := internal.PrepareValueForEncode("ProtoTimestampEncode", val, internal.TProtoTimestamp)
	if err != nil {
		return err
	}

	ts := val.Interface().(*timestamppb.Timestamp)
	if err := ts.CheckValid(); err != nil {
		return err
	}

	t := primitive.NewDateTimeFromTime(ts.AsTime())
	rv := reflect.ValueOf(&t).Elem()

	encoder, err := r.LookupEncoder(rv.Type())
	if err != nil {
		return err
	}

	return encoder.EncodeValue(bsoncodec.EncodeContext{Registry: r.Registry}, vw, rv)
}

// DecodeValue handles decoding of the `google.protobuf.Timestamp` message.
func (sc *ProtoTimestampCodec) DecodeValue(r bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	val, err := internal.PrepareValueForDecode("ProtoTimestampDecodeValue", val, internal.TProtoTimestamp)
	if err != nil {
		return err
	}

	ts := val.Interface().(*timestamppb.Timestamp)

	t := primitive.DateTime(0)
	rv := reflect.ValueOf(&t).Elem()

	decoder, err := r.LookupDecoder(rv.Type())

	if err := decoder.DecodeValue(bsoncodec.DecodeContext{Registry: r.Registry, Ancestor: r.Ancestor}, vr, rv); err != nil {
		return err
	}

	tc := t.Time()
	ts.Seconds = tc.Unix()
	ts.Nanos = int32(tc.Nanosecond())

	if err := ts.CheckValid(); err != nil {
		return err
	}

	return nil
}
