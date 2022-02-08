package protobson

import (
	"github.com/rdmrcv/go-protom/pkg/internal"
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

	rb.
		RegisterHookEncoder(internal.TProtoEnum, bsoncodec.ValueEncoderFunc(benumcodec.EncodeValue)).
		RegisterHookDecoder(internal.TProtoEnum, bsoncodec.ValueDecoderFunc(benumcodec.DecodeValue)).
		RegisterHookEncoder(internal.TProtoMessage, bsoncodec.ValueEncoderFunc(bpbcodec.EncodeValue)).
		RegisterHookDecoder(internal.TProtoMessage, bsoncodec.ValueDecoderFunc(bpbcodec.DecodeValue))

	return rb
}
