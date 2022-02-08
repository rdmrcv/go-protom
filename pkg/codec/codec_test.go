package codec

import (
	"testing"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/bson"
)

func TestProtoBsonCodec_EncodeValue(t *testing.T) {
	msg := internal.Msg{
		Field:   &internal.Embed{Text: "test"},
		Variant: &internal.Msg_A{A: "variant-a"},
	}

	bts, err := bson.MarshalWithRegistry(BsonPBRegistry, msg)
	if err != nil {
		panic(err)
	}

	outMsg := &internal.Msg{}
	if err := bson.UnmarshalWithRegistry(BsonPBRegistry, bts, outMsg); err != nil {
		panic(err)
	}

	t.Logf("%+v", outMsg)
}
