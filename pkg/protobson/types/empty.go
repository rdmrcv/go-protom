package types

import (
	"errors"
	"reflect"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
)

var (
	ErrEmptyIsNotEmpty = errors.New("empty object expected to be empty but it is not")
)

// ProtoEmptyCodec is the Codec used for google.protobuf.Empty type.
type ProtoEmptyCodec struct {
}

var _ bsoncodec.ValueEncoder = &ProtoEmptyCodec{}
var _ bsoncodec.ValueDecoder = &ProtoEmptyCodec{}

// NewProtoEmptyCodec returns a ProtoEmptyCodec that uses p for struct tag parsing.
func NewProtoEmptyCodec() *ProtoEmptyCodec {
	return &ProtoEmptyCodec{}
}

// EncodeValue handles encoding of the `google.protobuf.Empty` message.
func (sc *ProtoEmptyCodec) EncodeValue(r bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	val, err := internal.PrepareValueForEncode("ProtoEmptyEncode", val, internal.TProtoEmpty)
	if err != nil {
		return err
	}

	dw, err := vw.WriteDocument()
	if err != nil {
		return err
	}

	return dw.WriteDocumentEnd()
}

// DecodeValue handles decoding of the `google.protobuf.Empty` message.
func (sc *ProtoEmptyCodec) DecodeValue(r bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	val, err := internal.PrepareValueForDecode("ProtoEmptyDecodeValue", val, internal.TProtoEmpty)
	if err != nil {
		return err
	}

	dr, err := vr.ReadDocument()
	if err != nil {
		return err
	}

	if _, _, err := dr.ReadElement(); err == nil || !errors.Is(err, bsonrw.ErrEOD) {
		return ErrEmptyIsNotEmpty
	}

	return nil
}
