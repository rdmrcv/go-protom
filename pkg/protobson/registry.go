package protobson

import (
	"github.com/rdmrcv/go-protom/pkg/internal"
	"github.com/rdmrcv/go-protom/pkg/protobson/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
)

// BsonPBRegistry that provide protojson compatible BSON encoding.
var BsonPBRegistry = NewBsonPBRegistryBuilder().Build()

func NewBsonPBRegistryBuilder() *bsoncodec.RegistryBuilder {
	rb := bsoncodec.NewRegistryBuilder()
	bsoncodec.DefaultValueEncoders{}.RegisterDefaultEncoders(rb)
	bsoncodec.DefaultValueDecoders{}.RegisterDefaultDecoders(rb)
	bson.PrimitiveCodecs{}.RegisterPrimitiveCodecs(rb)

	bpbcodec := NewProtoMessageCodec()
	benumcodec := NewProtoEnumCodec()
	bwrapperscodec := types.NewProtoWrappersCodec()
	bdurationcodec := types.NewProtoDurationCodec()
	btimestampcodec := types.NewProtoTimestampCodec()
	bstructcodec := types.NewProtoStructCodec()
	bstructvaluecodec := types.NewProtoStructValueCodec()
	bstructlistvaluecodec := types.NewProtoStructListValueCodec()
	bemptycodec := types.NewProtoEmptyCodec()

	rb.
		RegisterTypeEncoder(internal.TProtoEmpty, bsoncodec.ValueEncoderFunc(bemptycodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoEmpty, bsoncodec.ValueDecoderFunc(bemptycodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoStruct, bsoncodec.ValueEncoderFunc(bstructcodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoStruct, bsoncodec.ValueDecoderFunc(bstructcodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoStructValue, bsoncodec.ValueEncoderFunc(bstructvaluecodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoStructValue, bsoncodec.ValueDecoderFunc(bstructvaluecodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoStructListValue, bsoncodec.ValueEncoderFunc(bstructlistvaluecodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoStructListValue, bsoncodec.ValueDecoderFunc(bstructlistvaluecodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoDuration, bsoncodec.ValueEncoderFunc(bdurationcodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoDuration, bsoncodec.ValueDecoderFunc(bdurationcodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoTimestamp, bsoncodec.ValueEncoderFunc(btimestampcodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoTimestamp, bsoncodec.ValueDecoderFunc(btimestampcodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoWrappedString, bsoncodec.ValueEncoderFunc(bwrapperscodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoWrappedString, bsoncodec.ValueDecoderFunc(bwrapperscodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoWrappedBytes, bsoncodec.ValueEncoderFunc(bwrapperscodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoWrappedBytes, bsoncodec.ValueDecoderFunc(bwrapperscodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoWrappedBool, bsoncodec.ValueEncoderFunc(bwrapperscodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoWrappedBool, bsoncodec.ValueDecoderFunc(bwrapperscodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoWrappedUInt32, bsoncodec.ValueEncoderFunc(bwrapperscodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoWrappedUInt32, bsoncodec.ValueDecoderFunc(bwrapperscodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoWrappedInt32, bsoncodec.ValueEncoderFunc(bwrapperscodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoWrappedInt32, bsoncodec.ValueDecoderFunc(bwrapperscodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoWrappedUInt64, bsoncodec.ValueEncoderFunc(bwrapperscodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoWrappedUInt64, bsoncodec.ValueDecoderFunc(bwrapperscodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoWrappedInt64, bsoncodec.ValueEncoderFunc(bwrapperscodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoWrappedInt64, bsoncodec.ValueDecoderFunc(bwrapperscodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoWrappedFloat, bsoncodec.ValueEncoderFunc(bwrapperscodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoWrappedFloat, bsoncodec.ValueDecoderFunc(bwrapperscodec.DecodeValue)).
		RegisterTypeEncoder(internal.TProtoWrappedDouble, bsoncodec.ValueEncoderFunc(bwrapperscodec.EncodeValue)).
		RegisterTypeDecoder(internal.TProtoWrappedDouble, bsoncodec.ValueDecoderFunc(bwrapperscodec.DecodeValue)).
		RegisterHookEncoder(internal.TProtoEnum, bsoncodec.ValueEncoderFunc(benumcodec.EncodeValue)).
		RegisterHookDecoder(internal.TProtoEnum, bsoncodec.ValueDecoderFunc(benumcodec.DecodeValue)).
		RegisterHookEncoder(internal.TProtoMessage, bsoncodec.ValueEncoderFunc(bpbcodec.EncodeValue)).
		RegisterHookDecoder(internal.TProtoMessage, bsoncodec.ValueDecoderFunc(bpbcodec.DecodeValue))

	return rb
}
