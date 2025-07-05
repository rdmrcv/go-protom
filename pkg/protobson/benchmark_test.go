package protobson

import (
	"bytes"
	"testing"
	"time"

	"github.com/rdmrcv/go-protom/v2/pkg/internal"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// createWKTMessage creates a test WKT message with all fields populated
func createWKTMessage() *internal.WKT {
	sv, _ := structpb.NewStruct(
		map[string]interface{}{
			"test": "some value",
		},
	)

	// Use a fixed time to ensure consistent benchmarks
	tm := bson.NewDateTimeFromTime(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)).Time()

	return &internal.WKT{
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
}

// BenchmarkProtoBsonMarshal benchmarks the marshaling of a WKT message using protobson
func BenchmarkProtoBsonMarshal(b *testing.B) {
	msg := createWKTMessage()
	reg := NewBsonPBRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := marshalWithRegistry(reg, msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProtoBsonUnmarshal benchmarks the unmarshaling of a WKT message using protobson
func BenchmarkProtoBsonUnmarshal(b *testing.B) {
	msg := createWKTMessage()
	reg := NewBsonPBRegistry()

	bts, err := marshalWithRegistry(reg, msg)
	if err != nil {
		b.Fatal(err)
	}

	dec := bson.NewDecoder(nil)
	dec.SetRegistry(reg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outMsg := &internal.WKT{}
		dec.Reset(bson.NewDocumentReader(bytes.NewReader(bts)))

		if err := dec.Decode(outMsg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProtoJsonMarshal benchmarks the marshaling of a WKT message using protojson
func BenchmarkProtoJsonMarshal(b *testing.B) {
	msg := createWKTMessage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := protojson.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProtoJsonUnmarshal benchmarks the unmarshaling of a WKT message using protojson
func BenchmarkProtoJsonUnmarshal(b *testing.B) {
	msg := createWKTMessage()

	bts, err := protojson.Marshal(msg)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outMsg := &internal.WKT{}
		if err := protojson.Unmarshal(bts, outMsg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProtoBsonRoundTrip benchmarks a complete round trip (marshal + unmarshal) using protobson
func BenchmarkProtoBsonRoundTrip(b *testing.B) {
	msg := createWKTMessage()
	reg := NewBsonPBRegistry()

	dec := bson.NewDecoder(nil)
	dec.SetRegistry(reg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bts, err := marshalWithRegistry(reg, msg)
		if err != nil {
			b.Fatal(err)
		}

		outMsg := &internal.WKT{}
		dec.Reset(bson.NewDocumentReader(bytes.NewReader(bts)))

		if err := dec.Decode(outMsg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProtoJsonRoundTrip benchmarks a complete round trip (marshal + unmarshal) using protojson
func BenchmarkProtoJsonRoundTrip(b *testing.B) {
	msg := createWKTMessage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bts, err := protojson.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}

		outMsg := &internal.WKT{}
		if err := protojson.Unmarshal(bts, outMsg); err != nil {
			b.Fatal(err)
		}
	}
}
