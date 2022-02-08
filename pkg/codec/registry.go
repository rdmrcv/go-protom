package codec

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

	bpbcodec := NewProtoBsonCodec()

	rb.
		RegisterHookEncoder(internal.TProto, bsoncodec.ValueEncoderFunc(bpbcodec.EncodeValue)).
		RegisterHookDecoder(internal.TProto, bsoncodec.ValueDecoderFunc(bpbcodec.DecodeValue))

	return rb
}
