package protobson

import (
	"github.com/rdmrcv/go-protom/pkg/internal"
	"github.com/rdmrcv/go-protom/pkg/protobson/types"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// BsonPBRegistry that provide protojson compatible BSON encoding.
var BsonPBRegistry = NewBsonPBRegistry()

func NewBsonPBRegistry() *bson.Registry {
	reg := bson.NewRegistry()

	bpbcodec := NewProtoMessageCodec()
	benumcodec := NewProtoEnumCodec()
	bwrapperscodec := types.NewProtoWrappersCodec()
	bdurationcodec := types.NewProtoDurationCodec()
	btimestampcodec := types.NewProtoTimestampCodec()
	bstructcodec := types.NewProtoStructCodec()
	bstructvaluecodec := types.NewProtoStructValueCodec()
	bstructlistvaluecodec := types.NewProtoStructListValueCodec()
	bemptycodec := types.NewProtoEmptyCodec()

	reg.RegisterTypeEncoder(internal.TProtoEmpty, bson.ValueEncoderFunc(bemptycodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoEmpty, bson.ValueDecoderFunc(bemptycodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoStruct, bson.ValueEncoderFunc(bstructcodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoStruct, bson.ValueDecoderFunc(bstructcodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoStructValue, bson.ValueEncoderFunc(bstructvaluecodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoStructValue, bson.ValueDecoderFunc(bstructvaluecodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoStructListValue, bson.ValueEncoderFunc(bstructlistvaluecodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoStructListValue, bson.ValueDecoderFunc(bstructlistvaluecodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoDuration, bson.ValueEncoderFunc(bdurationcodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoDuration, bson.ValueDecoderFunc(bdurationcodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoTimestamp, bson.ValueEncoderFunc(btimestampcodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoTimestamp, bson.ValueDecoderFunc(btimestampcodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoWrappedString, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoWrappedString, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoWrappedBytes, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoWrappedBytes, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoWrappedBool, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoWrappedBool, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoWrappedUInt32, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoWrappedUInt32, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoWrappedInt32, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoWrappedInt32, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoWrappedUInt64, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoWrappedUInt64, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoWrappedInt64, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoWrappedInt64, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoWrappedFloat, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoWrappedFloat, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	reg.RegisterTypeEncoder(internal.TProtoWrappedDouble, bson.ValueEncoderFunc(bwrapperscodec.EncodeValue))
	reg.RegisterTypeDecoder(internal.TProtoWrappedDouble, bson.ValueDecoderFunc(bwrapperscodec.DecodeValue))
	reg.RegisterInterfaceEncoder(internal.TProtoEnum, bson.ValueEncoderFunc(benumcodec.EncodeValue))
	reg.RegisterInterfaceDecoder(internal.TProtoEnum, bson.ValueDecoderFunc(benumcodec.DecodeValue))
	reg.RegisterInterfaceEncoder(internal.TProtoMessage, bson.ValueEncoderFunc(bpbcodec.EncodeValue))
	reg.RegisterInterfaceDecoder(internal.TProtoMessage, bson.ValueDecoderFunc(bpbcodec.DecodeValue))

	return reg
}
