package protobson

import (
	"testing"
	"time"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestProtoBsonCodec(t *testing.T) {
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
		t.Fatalf("messages is not equal to protojson")
	}

	if !proto.Equal(outMsg, msg) {
		t.Fatalf("messages is not equal to original")
	}
}

func TestWKT(t *testing.T) {
	sv, err := structpb.NewStruct(map[string]interface{}{
		"test": "some value",
	})

	if err != nil {
		panic(err)
	}

	// time in protojson and protobson differs from each other.
	// protobson uses bson datetime which loose precision.
	// To compare values after decode we should prune that extra precision bytes
	// before encode values via protojson.
	tm := primitive.NewDateTimeFromTime(time.Now()).Time()

	var msg = &internal.WKT{
		EmptyField:    &emptypb.Empty{},
		RepEmptyField: []*emptypb.Empty{{}, {}},

		DurationField: durationpb.New(time.Second * 3),
		RepDurationField: []*durationpb.Duration{
			durationpb.New(time.Second * 5),
			durationpb.New(time.Second * 6),
		},

		TimestampField: timestamppb.New(tm),
		RepTimestampField: []*timestamppb.Timestamp{
			timestamppb.New(tm),
		},

		StructField:    sv,
		RepStructField: []*structpb.Struct{sv},

		ValueField:    sv.Fields["test"],
		RepValueField: []*structpb.Value{sv.Fields["test"]},

		BoolField: wrapperspb.Bool(true),
		RepBoolField: []*wrapperspb.BoolValue{
			wrapperspb.Bool(true),
		},

		BytesField:    wrapperspb.Bytes([]byte{1, 2, 3}),
		RepBytesField: []*wrapperspb.BytesValue{wrapperspb.Bytes([]byte{4, 5, 6})},

		StringField: wrapperspb.String("test"),
		RepStringField: []*wrapperspb.StringValue{
			wrapperspb.String("test2"),
		},

		Int32Field: wrapperspb.Int32(6),
		RepInt32Field: []*wrapperspb.Int32Value{
			wrapperspb.Int32(32),
		},

		UInt32Field: wrapperspb.UInt32(6),
		RepUInt32Field: []*wrapperspb.UInt32Value{
			wrapperspb.UInt32(32),
		},

		Int64Field: wrapperspb.Int64(6),
		RepInt64Field: []*wrapperspb.Int64Value{
			wrapperspb.Int64(32),
		},

		UInt64Field: wrapperspb.UInt64(6),
		RepUInt64Field: []*wrapperspb.UInt64Value{
			wrapperspb.UInt64(32),
		},

		FloatField: wrapperspb.Float(333.3),
		RepFloatField: []*wrapperspb.FloatValue{
			wrapperspb.Float(2222.33),
		},

		DoubleField: wrapperspb.Double(333.3),
		RepDoubleField: []*wrapperspb.DoubleValue{
			wrapperspb.Double(2222.33),
		},
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

	outMsg, outJsonMsg := &internal.WKT{}, &internal.WKT{}

	if err := bson.UnmarshalWithRegistry(reg, bts, outMsg); err != nil {
		panic(err)
	}

	if err := protojson.Unmarshal(btsJson, outJsonMsg); err != nil {
		panic(err)
	}

	if !proto.Equal(outMsg, msg) {
		t.Fatalf("messages is not equal to original")
	}

	if !proto.Equal(outMsg, outJsonMsg) {
		t.Fatalf("messages is not equal to protojson")
	}
}
