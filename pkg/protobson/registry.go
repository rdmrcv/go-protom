package protobson

import (
	"github.com/rdmrcv/go-protom/pkg/internal"
	"github.com/rdmrcv/go-protom/pkg/protobson/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// BsonPBRegistry that provide protojson compatible BSON encoding.
var BsonPBRegistry = NewBsonPBRegistry()

func NewBsonPBRegistry() *bson.Registry {
	rb := bson.NewRegistry()

	bpbcodec := NewProtoMessageCodec()
	benumcodec := NewProtoEnumCodec()
	bwrapperscodec := types.NewProtoWrappersCodec()
	bdurationcodec := types.NewProtoDurationCodec()
	btimestampcodec := types.NewProtoTimestampCodec()
	bstructcodec := types.NewProtoStructCodec()
	bstructvaluecodec := types.NewProtoStructValueCodec()
	bstructlistvaluecodec := types.NewProtoStructListValueCodec()
	bemptycodec := types.NewProtoEmptyCodec()

	rb.RegisterTypeEncoder(internal.TProtoEmpty, bson.ValueEncoderFunc(bemptycodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoEmpty, bson.ValueDecoderFunc(bemptycodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoStruct, bson.ValueEncoderFunc(bstructcodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoStruct, bson.ValueDecoderFunc(bstructcodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoStructValue, bson.ValueEncoderFunc(bstructvaluecodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoStructValue, bson.ValueDecoderFunc(bstructvaluecodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoStructListValue, bson.ValueEncoderFunc(bstructlistvaluecodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoStructListValue, bson.ValueDecoderFunc(bstructlistvaluecodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoDuration, bson.ValueEncoderFunc(bdurationcodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoDuration, bson.ValueDecoderFunc(bdurationcodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoTimestamp, bson.ValueEncoderFunc(btimestampcodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoTimestamp, bson.ValueDecoderFunc(btimestampcodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoWrappedString, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoWrappedString, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoWrappedBytes, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoWrappedBytes, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoWrappedBool, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoWrappedBool, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoWrappedUInt32, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoWrappedUInt32, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoWrappedInt32, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoWrappedInt32, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoWrappedUInt64, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoWrappedUInt64, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoWrappedInt64, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoWrappedInt64, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoWrappedFloat, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoWrappedFloat, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	rb.RegisterTypeEncoder(internal.TProtoWrappedDouble, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	rb.RegisterTypeDecoder(internal.TProtoWrappedDouble, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	rb.RegisterInterfaceEncoder(internal.TProtoEnum, bson.ValueEncoderFunc(benumcodec.EncodeValue))
	rb.RegisterInterfaceDecoder(internal.TProtoEnum, bson.ValueDecoderFunc(benumcodec.DecodeValue))
	rb.RegisterInterfaceEncoder(internal.TProtoMessage, bson.ValueEncoderFunc(bpbcodec.EncodeValue))
	rb.RegisterInterfaceDecoder(internal.TProtoMessage, bson.ValueDecoderFunc(bpbcodec.DecodeValue))

	return rb
}
