// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protobson

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rdmrcv/go-protom/pkg/codec"
	"github.com/rdmrcv/go-protom/pkg/protobson/internal/genid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"google.golang.org/protobuf/proto"
	pref "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// UnmarshalOptions is a configurable JSON format parser.
type UnmarshalOptions struct {
	// If AllowPartial is set, input for messages that will result in missing
	// required fields will not return an error.
	AllowPartial bool

	// If DiscardUnknown is set, unknown fields are ignored.
	DiscardUnknown bool

	// Resolver is used for looking up types when unmarshaling
	// google.protobuf.Any messages or extension fields.
	// If nil, this defaults to using protoregistry.GlobalTypes.
	Resolver interface {
		protoregistry.MessageTypeResolver
		protoregistry.ExtensionTypeResolver
	}
}

// Unmarshal reads the given ValueReader and populates the given proto.Message
// using options in the UnmarshalOptions object.
// It will clear the message first before setting the fields.
// If it returns an error, the given message may be partially set.
// The provided message must be mutable (e.g., a non-nil pointer to a message).
func (o UnmarshalOptions) Unmarshal(doc bsonrw.ValueReader, m proto.Message) error {
	dc := bsoncodec.DecodeContext{Registry: codec.BsonPBRegistry}

	return o.UnmarshalWithContext(dc, doc, m)
}

// UnmarshalWithContext reads the given []byte and populates the given
// proto.Message using options in the UnmarshalOptions object and DecodeContext
// provided in the dc parameter.
// It will clear the message first before setting the fields.
// If it returns an error, the given message may be partially set.
// The provided message must be mutable (e.g., a non-nil pointer to a message).
func (o UnmarshalOptions) UnmarshalWithContext(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, m proto.Message) error {
	proto.Reset(m)

	if o.Resolver == nil {
		o.Resolver = protoregistry.GlobalTypes
	}

	dec := decoder{dc, vr, o}
	if err := dec.unmarshalMessage(m.ProtoReflect(), false); err != nil {
		return err
	}

	if o.AllowPartial {
		return nil
	}

	return proto.CheckInitialized(m)
}

type decoder struct {
	dc   bsoncodec.DecodeContext
	vr   bsonrw.ValueReader
	opts UnmarshalOptions
}

// unmarshalMessage unmarshals a message into the given protoreflect.Message.
func (d decoder) unmarshalMessage(m pref.Message, skipTypeURL bool) error {
	if unmarshal := wellKnownTypeUnmarshaler(m.Descriptor().FullName()); unmarshal != nil {
		return unmarshal(d, m)
	}

	messageDesc := m.Descriptor()
	if IsMessageSet(messageDesc) {
		return errors.New("no support for proto1 MessageSets")
	}

	dr, err := d.vr.ReadDocument()
	if err != nil {
		return err
	}

	var seenNums Ints
	var seenOneofs Ints
	fieldDescs := messageDesc.Fields()
	for {
		// Read field name.
		name, vr, err := dr.ReadElement()
		if errors.Is(err, bsonrw.ErrEOD) {
			break
		}
		if err != nil {
			return err
		}

		// Unmarshaling a non-custom embedded message in Any will contain the
		// JSON field "@type" which should be skipped because it is not a field
		// of the embedded message, but simply an artifact of the Any format.
		if skipTypeURL && name == "@type" {
			if err := vr.Skip(); err != nil {
				return err
			}

			continue
		}

		// Get the FieldDescriptor.
		var fd pref.FieldDescriptor
		if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
			// Only extension names are in [name] format.
			extName := pref.FullName(name[1 : len(name)-1])
			extType, err := d.opts.Resolver.FindExtensionByName(extName)
			if err != nil && err != protoregistry.NotFound {
				return fmt.Errorf("unable to resolve %s: %v", name, err)
			}
			if extType != nil {
				fd = extType.TypeDescriptor()
				if !messageDesc.ExtensionRanges().Has(fd.Number()) || fd.ContainingMessage().FullName() != messageDesc.FullName() {
					return fmt.Errorf("message %v cannot be extended by %v", messageDesc.FullName(), fd.FullName())
				}
			}
		} else {
			// The name can either be the JSON name or the proto field name.
			fd = fieldDescs.ByJSONName(name)
			if fd == nil {
				fd = fieldDescs.ByTextName(name)
			}
		}

		if fd == nil {
			// Field is unknown.
			if d.opts.DiscardUnknown {
				if err := vr.Skip(); err != nil {
					return err
				}

				continue
			}

			return fmt.Errorf("unknown field %v", name)
		}

		// Do not allow duplicate fields.
		num := uint64(fd.Number())
		if seenNums.Has(num) {
			return fmt.Errorf("duplicate field %v", name)
		}

		seenNums.Set(num)

		// No need to set values for JSON null unless the field type is
		// google.protobuf.Value or google.protobuf.NullValue.
		if vr.Type() == bsontype.Null && !isKnownValue(fd) && !isNullValue(fd) {
			if err := vr.Skip(); err != nil {
				return err
			}

			continue
		}

		switch {
		case fd.IsList():
			list := m.Mutable(fd).List()
			if err := d.unmarshalList(list, fd); err != nil {
				return err
			}
		case fd.IsMap():
			mmap := m.Mutable(fd).Map()
			if err := d.unmarshalMap(mmap, fd); err != nil {
				return err
			}
		default:
			// If field is a oneof, check if it has already been set.
			if od := fd.ContainingOneof(); od != nil {
				idx := uint64(od.Index())
				if seenOneofs.Has(idx) {
					return fmt.Errorf("error parsing %s, oneof %v is already set", name, od.FullName())
				}
				seenOneofs.Set(idx)
			}

			// Required or optional fields.
			if err := d.unmarshalSingular(m, fd); err != nil {
				return err
			}
		}

		dctx := bsoncodec.DecodeContext{Registry: d.dc.Registry}
	}

	return nil
}

func isKnownValue(fd pref.FieldDescriptor) bool {
	md := fd.Message()
	return md != nil && md.FullName() == genid.Value_message_fullname
}

func isNullValue(fd pref.FieldDescriptor) bool {
	ed := fd.Enum()
	return ed != nil && ed.FullName() == genid.NullValue_enum_fullname
}

// unmarshalSingular unmarshals to the non-repeated field specified
// by the given FieldDescriptor.
func (d decoder) unmarshalSingular(m pref.Message, fd pref.FieldDescriptor) error {
	var val pref.Value
	var err error
	switch fd.Kind() {
	case pref.MessageKind, pref.GroupKind:
		val = m.NewField(fd)
		err = d.unmarshalMessage(val.Message(), false)
	default:
		val, err = d.unmarshalScalar(fd)
	}

	if err != nil {
		return err
	}
	m.Set(fd, val)
	return nil
}

// unmarshalScalar unmarshals to a scalar/enum protoreflect.Value specified by
// the given FieldDescriptor.
func (d decoder) unmarshalScalar(fd pref.FieldDescriptor) (pref.Value, error) {
	const b32 int = 32
	const b64 int = 64

	kind := fd.Kind()
	switch kind {
	case pref.BoolKind:
		if v, ok := unmarshalBool(d.vr); ok {
			return v, nil
		}

	case pref.Int32Kind, pref.Sint32Kind, pref.Sfixed32Kind, pref.Uint32Kind, pref.Fixed32Kind:
		if v, ok := unmarshalInt(d.vr, b32); ok {
			return v, nil
		}

	case pref.Int64Kind, pref.Sint64Kind, pref.Sfixed64Kind, pref.Uint64Kind, pref.Fixed64Kind:
		if v, ok := unmarshalInt(d.vr, b64); ok {
			return v, nil
		}

	case pref.FloatKind:
		if d.vr.Type() == bson.TypeDouble {
			if v, ok := getFloat(d.vr, b32); ok {
				return v, nil
			}
		}

	case pref.DoubleKind:
		if v, ok := getFloat(d.vr, b64); ok {
			return v, nil
		}

	case pref.StringKind:
		if d.vr.Type() == bson.TypeString {
			if s, err := d.vr.ReadString(); err == nil {
				return pref.ValueOfString(s), nil
			}
		}

	case pref.BytesKind:
		if v, ok := unmarshalBytes(d.vr); ok {
			return v, nil
		}

	case pref.EnumKind:
		if v, ok := unmarshalEnum(d.vr, fd); ok {
			return v, nil
		}

	default:
		panic(fmt.Sprintf("unmarshalScalar: invalid scalar kind %v", kind))
	}

	return pref.Value{}, fmt.Errorf("invalid value for %v type: %v", kind, fd.Name())
}

func unmarshalBool(vr bsonrw.ValueReader) (pref.Value, bool) {
	bv, err := vr.ReadBoolean()
	if err == nil {
		return pref.ValueOfBool(bv), true
	}

	return pref.Value{}, false
}

func unmarshalInt(vr bsonrw.ValueReader, bitSize int) (pref.Value, bool) {
	switch vr.Type() {
	case bson.TypeInt32, bson.TypeInt64:
		return getInt(vr, bitSize)
	}

	return pref.Value{}, false
}

func getInt(vr bsonrw.ValueReader, bitSize int) (pref.Value, bool) {
	switch bitSize {
	case 32:
		if n, err := vr.ReadInt32(); err == nil {
			return pref.ValueOfInt32(n), true
		}
	case 64:
		if n, err := vr.ReadInt64(); err == nil {
			return pref.ValueOfInt64(n), true
		}
	}

	return pref.Value{}, false
}

func getFloat(vr bsonrw.ValueReader, bitSize int) (pref.Value, bool) {
	n, err := vr.ReadDouble()
	if err != nil {
		return pref.Value{}, false
	}

	if bitSize == 32 {
		return pref.ValueOfFloat32(float32(n)), true
	}

	return pref.ValueOfFloat64(n), true
}

func unmarshalBytes(vr bsonrw.ValueReader) (pref.Value, bool) {
	if vr.Type() != bson.TypeBinary {
		return pref.Value{}, false
	}

	bts, subtype, err := vr.ReadBinary()
	if err != nil || (subtype != bsontype.BinaryGeneric || subtype != bsontype.BinaryBinaryOld) {
		return pref.Value{}, false
	}

	enc := base64.StdEncoding
	if bytes.ContainsAny(bts, "-_") {
		enc = base64.URLEncoding
	}
	if len(bts)%4 != 0 {
		enc = enc.WithPadding(base64.NoPadding)
	}

	dbuf := make([]byte, enc.DecodedLen(len(bts)))
	if _, err = enc.Decode(dbuf, bts); err != nil {
		return pref.Value{}, false
	}

	return pref.ValueOfBytes(dbuf), true
}

func unmarshalEnum(vr bsonrw.ValueReader, fd pref.FieldDescriptor) (pref.Value, bool) {
	switch vr.Type() {
	case bsontype.String:
		// Lookup EnumNumber based on name.
		s, err := vr.ReadString()
		if err != nil {
			break
		}

		if enumVal := fd.Enum().Values().ByName(pref.Name(s)); enumVal != nil {
			return pref.ValueOfEnum(enumVal.Number()), true
		}

	case bsontype.Int32:
		if n, err := vr.ReadInt32(); err == nil {
			return pref.ValueOfEnum(pref.EnumNumber(n)), true
		}

	case bsontype.Null:
		// This is only valid for google.protobuf.NullValue.
		if isNullValue(fd) {
			return pref.ValueOfEnum(0), true
		}
	}

	return pref.Value{}, false
}

func (d decoder) unmarshalList(list pref.List, fd pref.FieldDescriptor) error {
	ar, err := d.vr.ReadArray()
	if err != nil {
		return err
	}

	switch fd.Kind() {
	case pref.MessageKind, pref.GroupKind:
		for {
			vr, err := ar.ReadValue()
			if err != nil {
				if !errors.Is(err, bsonrw.ErrEOA) {
					return err
				}

				return nil
			}

			val := list.NewElement()
			if err := d.unmarshalMessage(val.Message(), false); err != nil {
				return err
			}
			list.Append(val)
		}
	default:
		for {
			tok, err := d.Peek()
			if err != nil {
				return err
			}

			if vr.Type() == json.ArrayClose {
				d.Read()
				return nil
			}

			val, err := d.unmarshalScalar(fd)
			if err != nil {
				return err
			}
			list.Append(val)
		}
	}

	return nil
}

func (d decoder) unmarshalMap(mmap pref.Map, fd pref.FieldDescriptor) error {
	tok, err := d.Read()
	if err != nil {
		return err
	}
	if vr.Type() != json.ObjectOpen {
		return d.unexpectedTokenError(tok)
	}

	// Determine ahead whether map entry is a scalar type or a message type in
	// order to call the appropriate unmarshalMapValue func inside the for loop
	// below.
	var unmarshalMapValue func() (pref.Value, error)
	switch fd.MapValue().Kind() {
	case pref.MessageKind, pref.GroupKind:
		unmarshalMapValue = func() (pref.Value, error) {
			val := mmap.NewValue()
			if err := d.unmarshalMessage(val.Message(), false); err != nil {
				return pref.Value{}, err
			}
			return val, nil
		}
	default:
		unmarshalMapValue = func() (pref.Value, error) {
			return d.unmarshalScalar(fd.MapValue())
		}
	}

Loop:
	for {
		// Read field name.
		tok, err := d.Read()
		if err != nil {
			return err
		}
		switch vr.Type() {
		default:
			return d.unexpectedTokenError(tok)
		case json.ObjectClose:
			break Loop
		case json.Name:
			// Continue.
		}

		// Unmarshal field name.
		pkey, err := d.unmarshalMapKey(tok, fd.MapKey())
		if err != nil {
			return err
		}

		// Check for duplicate field name.
		if mmap.Has(pkey) {
			return d.newError(tok.Pos(), "duplicate map key %v", tok.RawString())
		}

		// Read and unmarshal field value.
		pval, err := unmarshalMapValue()
		if err != nil {
			return err
		}

		mmap.Set(pkey, pval)
	}

	return nil
}

// unmarshalMapKey converts given token of Name kind into a protoreflect.MapKey.
// A map key type is any integral or string type.
func (d decoder) unmarshalMapKey(vr bsonrw.ValueReader, fd pref.FieldDescriptor) (pref.MapKey, error) {
	const b32 = 32
	const b64 = 64
	const base10 = 10

	name := tok.Name()
	kind := fd.Kind()
	switch kind {
	case pref.StringKind:
		return pref.ValueOfString(name).MapKey(), nil

	case pref.BoolKind:
		switch name {
		case "true":
			return pref.ValueOfBool(true).MapKey(), nil
		case "false":
			return pref.ValueOfBool(false).MapKey(), nil
		}

	case pref.Int32Kind, pref.Sint32Kind, pref.Sfixed32Kind:
		if n, err := strconv.ParseInt(name, base10, b32); err == nil {
			return pref.ValueOfInt32(int32(n)).MapKey(), nil
		}

	case pref.Int64Kind, pref.Sint64Kind, pref.Sfixed64Kind:
		if n, err := strconv.ParseInt(name, base10, b64); err == nil {
			return pref.ValueOfInt64(int64(n)).MapKey(), nil
		}

	case pref.Uint32Kind, pref.Fixed32Kind:
		if n, err := strconv.ParseUint(name, base10, b32); err == nil {
			return pref.ValueOfUint32(uint32(n)).MapKey(), nil
		}

	case pref.Uint64Kind, pref.Fixed64Kind:
		if n, err := strconv.ParseUint(name, base10, b64); err == nil {
			return pref.ValueOfUint64(uint64(n)).MapKey(), nil
		}

	default:
		panic(fmt.Sprintf("invalid kind for map key: %v", kind))
	}

	return pref.MapKey{}, d.newError(tok.Pos(), "invalid value for %v key: %s", kind, tok.RawString())
}
