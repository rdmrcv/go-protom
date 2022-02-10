package protobson

import (
	"reflect"
	"sync"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ProtoEnumCodec is the Codec used for struct values.
type ProtoEnumCodec struct {
	cache map[protoreflect.MessageType]*structDescription
	l     sync.RWMutex

	// UseEnumNumbers emits enum values as numbers.
	UseEnumNumbers bool
}

var _ bsoncodec.ValueEncoder = &ProtoEnumCodec{}
var _ bsoncodec.ValueDecoder = &ProtoEnumCodec{}

// NewProtoEnumCodec returns a ProtoEnumCodec that uses p for struct tag parsing.
func NewProtoEnumCodec() *ProtoEnumCodec {
	return &ProtoEnumCodec{
		cache: make(map[protoreflect.MessageType]*structDescription),
	}
}

// EncodeValue handles encoding generic struct types.
func (sc *ProtoEnumCodec) EncodeValue(r bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	val, err := internal.PrepareInterfaceForEncode("ProtoEnumEncodeValue", val, internal.TProtoEnum)
	if err != nil {
		return err
	}

	protoMsg := val.Interface().(protoreflect.Enum)

	if sc.UseEnumNumbers {
		return vw.WriteInt32(int32(protoMsg.Number()))
	} else {
		return vw.WriteString(string(protoMsg.Descriptor().Values().ByNumber(protoMsg.Number()).Name()))
	}
}

// DecodeValue implements the Codec interface.
// By default, map types in val will not be cleared. If a map has existing key/value pairs, it will be extended with the new ones from vr.
// For slices, the decoder will set the length of the slice to zero and append all elements. The underlying array will not be cleared.
func (sc *ProtoEnumCodec) DecodeValue(r bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	val, err := internal.PrepareInterfaceForDecode("ProtoEnumDecodeValue", val, internal.TProtoEnum)
	if err != nil {
		return err
	}

	protoMsg := val.Interface().(protoreflect.Enum)

	var num int32
	if !sc.UseEnumNumbers {
		val, err := vr.ReadString()
		if err != nil {
			return err
		}

		num = int32(protoMsg.Descriptor().Values().ByName(protoreflect.Name(val)).Number())
	} else {
		valReaden, err := vr.ReadInt32()
		if err != nil {
			return err
		}

		num = valReaden
	}

	rv := reflect.ValueOf(num)

	val.Set(rv.Convert(val.Type()))

	return nil
}
