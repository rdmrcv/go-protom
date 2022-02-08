package protobson

import (
	"testing"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestProtoBsonCodec_EncodeValue(t *testing.T) {
	var msg proto.Message = &internal.Msg{
		Field:   &internal.Embed{Text: "test"},
		Variant: &internal.Msg_C{C: &internal.Embed{Text: "variant-C"}},
		MapField: map[string]*internal.Embed{
			"test": {Text: "map-field"},
		},
		RepField: []*internal.Embed{
			{Text: "repeated-field"},
		},
		En:      internal.EnumType_ANOTHER,
		RepEnum: []internal.EnumType{internal.EnumType_TEST},
		MapEnum: map[uint32]internal.EnumType{100: internal.EnumType_ANOTHER},
	}

	reg := NewBsonPBRegistryBuilder().Build()

	bts, err := bson.MarshalWithRegistry(reg, msg)
	if err != nil {
		panic(err)
	}

	btsJson, err := protojson.Marshal(msg)
	if err != nil {
		panic(err)
	}

	outMsg, outJsonMsg := &internal.Msg{}, &internal.Msg{}

	if err := bson.UnmarshalWithRegistry(reg, bts, outMsg); err != nil {
		panic(err)
	}

	if err := protojson.Unmarshal(btsJson, outJsonMsg); err != nil {
		panic(err)
	}

	if !proto.Equal(outMsg, outJsonMsg) {
		t.Fatalf("messages is not equal")
	}
}
