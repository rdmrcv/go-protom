package types

import (
	"bytes"
	"reflect"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
)

// ProtoDurationCodec is the Codec used for google.protobuf.Duration type.
type ProtoDurationCodec struct {
}

var _ bsoncodec.ValueEncoder = &ProtoDurationCodec{}
var _ bsoncodec.ValueDecoder = &ProtoDurationCodec{}

// NewProtoDurationCodec returns a ProtoDurationCodec.
func NewProtoDurationCodec() *ProtoDurationCodec {
	return &ProtoDurationCodec{}
}

// EncodeValue handles encoding of the `google.protobuf.Duration` message.
func (sc *ProtoDurationCodec) EncodeValue(r bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	val, err := internal.PrepareValueForEncode("ProtoDurationEncode", val, internal.TProtoDuration)
	if err != nil {
		return err
	}

	protoMsg := val.Interface().(*durationpb.Duration)
	if err := protoMsg.CheckValid(); err != nil {
		return err
	}

	bts, err := protojson.Marshal(protoMsg)
	if err != nil {
		return err
	}

	// Use the protojson as-is because here we want a raw not bson-specific string.
	return vw.WriteString(string(bytes.Trim(bts, "\"")))
}

// DecodeValue handles decoding of the `google.protobuf.Duration` message.
func (sc *ProtoDurationCodec) DecodeValue(r bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	val, err := internal.PrepareValueForDecode("ProtoDurationDecodeValue", val, internal.TProtoDuration)
	if err != nil {
		return err
	}

	protoMsg := val.Interface().(*durationpb.Duration)

	d, err := vr.ReadString()
	if err != nil {
		return err
	}

	// Use the protojson as-is because here we have a raw not bson-specific string.
	return protojson.Unmarshal([]byte("\""+d+"\""), protoMsg)
}
