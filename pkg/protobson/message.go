package protobson

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var errEncodeInvalidValue = errors.New("cannot encode invalid element")
var errDecodeInvalidValue = errors.New("cannot decode invalid element")

var ErrUnknownField = errors.New("unknown field")

// ProtoMessageCodec is the Codec used for the protoreflect.ProtoMessage (aka proto.Message) values.
type ProtoMessageCodec struct {
	cache map[reflect.Type]*structDescription
	l     sync.RWMutex

	// AllowPartial allows messages that have missing required fields to marshal
	// without returning an error. If AllowPartial is false (the default),
	// Marshal will return error if there are any missing required fields.
	AllowMarshalPartial bool

	// If AllowPartial is set, input for messages that will result in missing
	// required fields will not return an error.
	AllowUnmarshalPartial bool

	// UseProtoNames uses proto field name instead of lowerCamelCase name in JSON
	// field names.
	UseProtoNames bool

	// UseEnumNumbers emits enum values as numbers.
	UseEnumNumbers bool

	// If DiscardUnknown is set, unknown fields are ignored.
	DiscardUnknown bool

	// Resolver is used for looking up types when expanding google.protobuf.Any
	// messages. If nil, this defaults to using protoregistry.GlobalTypes.
	Resolver interface {
		protoregistry.ExtensionTypeResolver
		protoregistry.MessageTypeResolver
	}
}

var _ bsoncodec.ValueEncoder = &ProtoMessageCodec{}
var _ bsoncodec.ValueDecoder = &ProtoMessageCodec{}

// NewProtoMessageCodec returns a ProtoMessageCodec that uses p for struct tag parsing.
func NewProtoMessageCodec() *ProtoMessageCodec {
	return &ProtoMessageCodec{
		cache: make(map[reflect.Type]*structDescription),
	}
}

// EncodeValue handles encoding generic struct types.
func (sc *ProtoMessageCodec) EncodeValue(r bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	val, err := internal.PrepareInterfaceForEncode("ProtoMessageEncodeValue", val, internal.TProtoMessage)
	if err != nil {
		return err
	}

	// If ValueMarshaler is implemented on a concrete type, make sure that val isn't a nil pointer
	if isImplementationNil(val, internal.TProtoMessage) {
		return vw.WriteNull()
	}

	protoMsg := val.Interface().(protoreflect.ProtoMessage)
	protoDesc := protoMsg.ProtoReflect()

	sd, err := sc.describeStruct(r.Registry, protoDesc, val.Elem().Type())
	if err != nil {
		return err
	}

	dw, err := vw.WriteDocument()
	if err != nil {
		return err
	}

	var rv protoreflect.Value
	var fd protoreflect.FieldDescriptor
	for _, desc := range sd.fl {
		fd = protoDesc.Descriptor().Fields().Get(desc.idx)
		rv = protoDesc.Get(fd)

		if desc.encoder == nil {
			return bsoncodec.ErrNoEncoder{Type: reflect.TypeOf(rv.Interface())}
		}

		encoder := desc.encoder

		if !protoDesc.Has(fd) {
			continue
		}

		vw2, err := dw.WriteDocumentElement(desc.name)
		if err != nil {
			return err
		}

		rv2, err := internal.GetRWValue(rv, protoMsg, desc.name)
		if err != nil {
			return err
		}

		ectx := bsoncodec.EncodeContext{Registry: r.Registry}
		err = encoder.EncodeValue(ectx, vw2, rv2)
		if err != nil {
			return err
		}
	}

	if !sc.AllowMarshalPartial {
		if err := proto.CheckInitialized(protoMsg); err != nil {
			return err
		}
	}

	return dw.WriteDocumentEnd()
}

// DecodeValue implements the Codec interface.
// By default, map types in val will not be cleared. If a map has existing key/value pairs, it will be extended with the new ones from vr.
// For slices, the decoder will set the length of the slice to zero and append all elements. The underlying array will not be cleared.
func (sc *ProtoMessageCodec) DecodeValue(r bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	val, err2 := internal.PrepareInterfaceForDecode("MessageDecodeValue", val, internal.TProtoMessage)
	if err2 != nil {
		return err2
	}

	protoMsg := val.Interface().(protoreflect.ProtoMessage)

	proto.Reset(protoMsg)

	protoDesc := protoMsg.ProtoReflect()

	sd, err := sc.describeStruct(r.Registry, protoDesc, val.Elem().Type())
	if err != nil {
		return err
	}

	dr, err := vr.ReadDocument()
	if err != nil {
		return err
	}

	var rv protoreflect.Value
	var fd protoreflect.FieldDescriptor
	var field reflect.Value
	for {
		name, vr, err := dr.ReadElement()
		if err == bsonrw.ErrEOD {
			break
		}
		if err != nil {
			return err
		}

		desc, exists := sd.fm[name]
		if !exists {
			// if the original name isn't found in the struct description, try again with the name in lowercase
			// this could match if a BSON tag isn't specified because by default, describeStruct lowercases all field
			// names
			desc, exists = sd.fm[strings.ToLower(name)]
		}

		if !exists {
			if sc.DiscardUnknown {
				// The encoding/json package requires a flag to return on error for non-existent fields.
				// This functionality seems appropriate for the struct codec.
				err = vr.Skip()
				if err != nil {
					return err
				}

				continue
			}

			return internal.NewDecodeError(desc.name, ErrUnknownField)
		}

		fd = protoDesc.Descriptor().Fields().Get(desc.idx)
		rv = protoDesc.NewField(fd)
		field, err = internal.GetRWValue(rv, protoMsg, desc.name)
		if err != nil {
			return err
		}

		dctx := bsoncodec.DecodeContext{Registry: r.Registry, Truncate: r.Truncate}
		if desc.decoder == nil {
			return internal.NewDecodeError(desc.name, bsoncodec.ErrNoDecoder{Type: field.Type()})
		}

		err = desc.decoder.DecodeValue(dctx, vr, field)
		if err != nil {
			return internal.NewDecodeError(desc.name, err)
		}

		switch rv.Interface().(type) {
		case protoreflect.Map, protoreflect.List, protoreflect.EnumNumber:
			// Field already point to struct enum
		default:
			decVal, err := internal.ReflectToProtoValue(field)
			if errors.Is(err, internal.ErrDirectSet) {
				continue
			}

			if err != nil {
				return internal.NewDecodeError(desc.name, err)
			}

			if en := fd.Enum(); en != nil && !sc.UseEnumNumbers {
				evd := en.Values().ByName(protoreflect.Name(decVal.String()))
				if evd == nil {
					return internal.NewDecodeError(desc.name, errDecodeInvalidValue)
				}

				protoDesc.Set(fd, protoreflect.ValueOf(evd.Number()))
			} else {
				protoDesc.Set(fd, decVal)
			}
		}
	}

	if !sc.AllowUnmarshalPartial {
		if err := proto.CheckInitialized(protoMsg); err != nil {
			return err
		}
	}

	return nil
}

type structDescription struct {
	fm map[string]fieldDescription
	fl []fieldDescription
}

type fieldDescription struct {
	name      string // BSON key name
	fieldName string // struct field name
	idx       int
	encoder   bsoncodec.ValueEncoder
	decoder   bsoncodec.ValueDecoder
}

type byIndex []fieldDescription

func (bi byIndex) Len() int { return len(bi) }

func (bi byIndex) Swap(i, j int) { bi[i], bi[j] = bi[j], bi[i] }

func (bi byIndex) Less(i, j int) bool {
	return bi[i].idx < bi[j].idx
}

func (sc *ProtoMessageCodec) describeStruct(r *bsoncodec.Registry, protoDesc protoreflect.Message, t reflect.Type) (*structDescription, error) {
	fieldsDesc := protoDesc.Descriptor().Fields()
	// We need to analyze the struct, including getting the tags, collecting
	// information about inlining, and create a map of the field name to the field.
	sc.l.RLock()
	ds, exists := sc.cache[t]
	sc.l.RUnlock()
	if exists {
		return ds, nil
	}

	numFields := fieldsDesc.Len()
	sd := &structDescription{
		fm: make(map[string]fieldDescription, numFields),
		fl: make([]fieldDescription, 0, numFields),
	}

	var sfType reflect.Type
	var sf protoreflect.FieldDescriptor
	var sfv protoreflect.Value
	var msg protoreflect.ProtoMessage

	for i := 0; i < numFields; i++ {
		sf = fieldsDesc.Get(i)
		sfv = protoDesc.Get(sf)

		switch sfv.Interface().(type) {
		case protoreflect.EnumNumber:
			sfType = internal.TProtoEnum
		case protoreflect.Map, protoreflect.List:
			// We cannot use stdlib reflection for most fields to allow
			// a protobuf machinery solve complex types like oneofs, but for
			// map and list we can do that because protobuf messages does not
			// support inline lists and maps, and we can be sure that the field
			// with the map or the list type will exists in the origin message.
			msg = protoDesc.Interface()
			sfType = reflect.ValueOf(msg).Elem().FieldByName(string(sf.Name())).Type()
		default:
			sfType = internal.ReflectIntoStd(sfv, sc.UseEnumNumbers)
		}

		encoder, err := r.LookupEncoder(sfType)
		if err != nil {
			encoder = nil
		}

		decoder, err := r.LookupDecoder(sfType)
		if err != nil {
			decoder = nil
		}

		description := fieldDescription{
			fieldName: sf.TextName(),
			idx:       i,
			encoder:   encoder,
			decoder:   decoder,
		}

		if sc.UseProtoNames {
			description.name = description.fieldName
		} else {
			description.name = sf.JSONName()
		}

		sd.fl = append(sd.fl, description)
		sd.fm[description.name] = description
	}

	sort.Sort(byIndex(sd.fl))

	sc.l.Lock()
	sc.cache[t] = sd
	sc.l.Unlock()

	return sd, nil
}

// isImplementationNil returns if val is a nil pointer and inter is implemented on a concrete type
func isImplementationNil(val reflect.Value, inter reflect.Type) bool {
	vt := val.Type()
	for vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	return vt.Implements(inter) && val.Kind() == reflect.Ptr && val.IsNil()
}
