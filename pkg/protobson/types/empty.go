package types

import (
	"errors"
	"reflect"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrEmptyIsNotEmpty = errors.New("empty object expected to be empty but it is not")
)

// ProtoEmptyCodec is the Codec used for google.protobuf.Empty type.
type ProtoEmptyCodec struct {
}

var _ bson.ValueEncoder = &ProtoEmptyCodec{}
var _ bson.ValueDecoder = &ProtoEmptyCodec{}

// NewProtoEmptyCodec returns a ProtoEmptyCodec that uses p for struct tag parsing.
func NewProtoEmptyCodec() *ProtoEmptyCodec {
	return &ProtoEmptyCodec{}
}

// EncodeValue handles encoding of the `google.protobuf.Empty` message.
func (sc *ProtoEmptyCodec) EncodeValue(r bson.EncodeContext, vw bson.ValueWriter, val reflect.Value) error {
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
func (sc *ProtoEmptyCodec) DecodeValue(r bson.DecodeContext, vr bson.ValueReader, val reflect.Value) error {
	val, err := internal.PrepareValueForDecode("ProtoEmptyDecodeValue", val, internal.TProtoEmpty)
	if err != nil {
		return err
	}

	dr, err := vr.ReadDocument()
	if err != nil {
		return err
	}

	if _, _, err := dr.ReadElement(); err == nil || !errors.Is(err, bson.ErrEOD) {
		return ErrEmptyIsNotEmpty
	}

	return nil
}
