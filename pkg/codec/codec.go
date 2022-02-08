package codec

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/rdmrcv/go-protom/pkg/internal"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"google.golang.org/protobuf/proto"
)

var errInvalidValue = errors.New("cannot encode invalid element")
var ErrUnknownField = errors.New("unknown field")

// DecodeError represents an error that occurs when unmarshalling BSON bytes into a native Go type.
type DecodeError struct {
	keys    []string
	wrapped error
}

// Unwrap returns the underlying error
func (de *DecodeError) Unwrap() error {
	return de.wrapped
}

// Error implements the error interface.
func (de *DecodeError) Error() string {
	// The keys are stored in reverse order because the de.keys slice is builtup while propagating the error up the
	// stack of BSON keys, so we call de.Keys(), which reverses them.
	keyPath := strings.Join(de.Keys(), ".")
	return fmt.Sprintf("error decoding key %s: %v", keyPath, de.wrapped)
}

// Keys returns the BSON key path that caused an error as a slice of strings. The keys in the slice are in top-down
// order. For example, if the document being unmarshalled was {a: {b: {c: 1}}} and the value for c was supposed to be
// a string, the keys slice will be ["a", "b", "c"].
func (de *DecodeError) Keys() []string {
	reversedKeys := make([]string, 0, len(de.keys))
	for idx := len(de.keys) - 1; idx >= 0; idx-- {
		reversedKeys = append(reversedKeys, de.keys[idx])
	}

	return reversedKeys
}

// Zeroer allows custom struct types to implement a report of zero
// state. All struct types that don't implement Zeroer or where IsZero
// returns false are considered to be not zero.
type Zeroer interface {
	IsZero() bool
}

// ProtoBsonCodec is the Codec used for struct values.
type ProtoBsonCodec struct {
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
}

var _ bsoncodec.ValueEncoder = &ProtoBsonCodec{}
var _ bsoncodec.ValueDecoder = &ProtoBsonCodec{}

// NewProtoBsonCodec returns a ProtoBsonCodec that uses p for struct tag parsing.
func NewProtoBsonCodec() (*ProtoBsonCodec, error) {
	codec := &ProtoBsonCodec{
		cache: make(map[reflect.Type]*structDescription),
	}

	return codec, nil
}

// EncodeValue handles encoding generic struct types.
func (sc *ProtoBsonCodec) EncodeValue(r bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	// Either val or a pointer to val must implement ValueMarshaler
	switch {
	case !val.IsValid():
		return bsoncodec.ValueEncoderError{Name: "BsonPBEncodeValue", Types: []reflect.Type{internal.TProto}, Received: val}
	case val.Type().Implements(internal.TProto):
		// If ValueMarshaler is implemented on a concrete type, make sure that val isn't a nil pointer
		if isImplementationNil(val, internal.TProto) {
			return vw.WriteNull()
		}
	case reflect.PtrTo(val.Type()).Implements(internal.TProto) && val.CanAddr():
		val = val.Addr()
	default:
		return bsoncodec.ValueEncoderError{Name: "BsonPBEncodeValue", Types: []reflect.Type{internal.TProto}, Received: val}
	}

	sd, err := sc.describeStruct(r.Registry, val)
	if err != nil {
		return err
	}

	dw, err := vw.WriteDocument()
	if err != nil {
		return err
	}

	var rv reflect.Value
	for _, desc := range sd.fl {
		if desc.inline == nil {
			rv = val.Field(desc.idx)
		} else {
			rv, err = fieldByIndexErr(val, desc.inline)
			if err != nil {
				continue
			}
		}

		desc.encoder, rv, err = lookupElementEncoder(r, desc.encoder, rv)

		if err != nil && err != errInvalidValue {
			return err
		}

		if err == errInvalidValue {
			continue
		}

		if desc.encoder == nil {
			return bsoncodec.ErrNoEncoder{Type: rv.Type()}
		}

		encoder := desc.encoder

		var isZero bool
		rvInterface := rv.Interface()
		if cz, ok := encoder.(bsoncodec.CodecZeroer); ok {
			isZero = cz.IsTypeZero(rvInterface)
		} else if rv.Kind() == reflect.Interface {
			// sc.isZero will not treat an interface rv as an interface, so we need to check for the zero interface separately.
			isZero = rv.IsNil()
		} else {
			isZero = sc.isZero(rvInterface)
		}

		if isZero {
			continue
		}

		vw2, err := dw.WriteDocumentElement(desc.name)
		if err != nil {
			return err
		}

		ectx := bsoncodec.EncodeContext{Registry: r.Registry}
		err = encoder.EncodeValue(ectx, vw2, rv)
		if err != nil {
			return err
		}
	}

	return dw.WriteDocumentEnd()
}

func newDecodeError(key string, original error) error {
	de, ok := original.(*DecodeError)
	if !ok {
		return &DecodeError{
			keys:    []string{key},
			wrapped: original,
		}
	}

	de.keys = append(de.keys, key)
	return de
}

// DecodeValue implements the Codec interface.
// By default, map types in val will not be cleared. If a map has existing key/value pairs, it will be extended with the new ones from vr.
// For slices, the decoder will set the length of the slice to zero and append all elements. The underlying array will not be cleared.
func (sc *ProtoBsonCodec) DecodeValue(r bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.IsValid() || (!val.Type().Implements(internal.TProto) && !reflect.PtrTo(val.Type()).Implements(internal.TProto)) {
		return bsoncodec.ValueDecoderError{Name: "BsonPBDecodeValue", Types: []reflect.Type{internal.TProto}, Received: val}
	}

	if val.Kind() == reflect.Ptr && val.IsNil() {
		if !val.CanSet() {
			return bsoncodec.ValueDecoderError{Name: "BsonPBDecodeValue", Types: []reflect.Type{internal.TProto}, Received: val}
		}
		val.Set(reflect.New(val.Type().Elem()))
	}

	if !val.Type().Implements(internal.TProto) {
		if !val.CanAddr() {
			return bsoncodec.ValueDecoderError{Name: "BsonPBDecodeValue", Types: []reflect.Type{internal.TProto}, Received: val}
		}
		val = val.Addr() // If they type doesn't implement the interface, a pointer to it must.
	}

	switch vrType := vr.Type(); vrType {
	case bsontype.Type(0), bsontype.EmbeddedDocument:
	case bsontype.Null:
		if err := vr.ReadNull(); err != nil {
			return err
		}

		val.Set(reflect.Zero(val.Type()))
		return nil
	case bsontype.Undefined:
		if err := vr.ReadUndefined(); err != nil {
			return err
		}

		val.Set(reflect.Zero(val.Type()))
		return nil
	default:
		return fmt.Errorf("cannot decode %v into a %s", vrType, val.Type())
	}

	sd, err := sc.describeStruct(r.Registry, val)
	if err != nil {
		return err
	}

	dr, err := vr.ReadDocument()
	if err != nil {
		return err
	}

	for {
		name, vr, err := dr.ReadElement()
		if err == bsonrw.ErrEOD {
			break
		}
		if err != nil {
			return err
		}

		fd, exists := sd.fm[name]
		if !exists {
			// if the original name isn't found in the struct description, try again with the name in lowercase
			// this could match if a BSON tag isn't specified because by default, describeStruct lowercases all field
			// names
			fd, exists = sd.fm[strings.ToLower(name)]
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

			return newDecodeError(fd.name, ErrUnknownField)
		}

		var field reflect.Value
		if fd.inline == nil {
			field = val.Field(fd.idx)
		} else {
			field, err = getInlineField(val, fd.inline)
			if err != nil {
				return err
			}
		}

		if !field.CanSet() { // Being settable is a super set of being addressable.
			innerErr := fmt.Errorf("field %v is not settable", field)
			return newDecodeError(fd.name, innerErr)
		}

		if field.Kind() == reflect.Ptr && field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}

		field = field.Addr()

		dctx := bsoncodec.DecodeContext{Registry: r.Registry, Truncate: r.Truncate}
		if fd.decoder == nil {
			return newDecodeError(fd.name, bsoncodec.ErrNoDecoder{Type: field.Elem().Type()})
		}

		err = fd.decoder.DecodeValue(dctx, vr, field.Elem())
		if err != nil {
			return newDecodeError(fd.name, err)
		}
	}

	return nil
}

func (sc *ProtoBsonCodec) isZero(i interface{}) bool {
	v := reflect.ValueOf(i)

	// check the value validity
	if !v.IsValid() {
		return true
	}

	if z, ok := v.Interface().(Zeroer); ok && (v.Kind() != reflect.Ptr || !v.IsNil()) {
		return z.IsZero()
	}

	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return false
}

type structDescription struct {
	fm     map[string]fieldDescription
	fl     []fieldDescription
	inline bool
}

type fieldDescription struct {
	name      string // BSON key name
	fieldName string // struct field name
	idx       int
	inline    []int
	encoder   bsoncodec.ValueEncoder
	decoder   bsoncodec.ValueDecoder
}

type byIndex []fieldDescription

func (bi byIndex) Len() int { return len(bi) }

func (bi byIndex) Swap(i, j int) { bi[i], bi[j] = bi[j], bi[i] }

func (bi byIndex) Less(i, j int) bool {
	// If a field is inlined, its index in the top level struct is stored at inline[0]
	iIdx, jIdx := bi[i].idx, bi[j].idx
	if len(bi[i].inline) > 0 {
		iIdx = bi[i].inline[0]
	}
	if len(bi[j].inline) > 0 {
		jIdx = bi[j].inline[0]
	}
	if iIdx != jIdx {
		return iIdx < jIdx
	}
	for k, biik := range bi[i].inline {
		if k >= len(bi[j].inline) {
			return false
		}
		if biik != bi[j].inline[k] {
			return biik < bi[j].inline[k]
		}
	}
	return len(bi[i].inline) < len(bi[j].inline)
}

func (sc *ProtoBsonCodec) describeStruct(r *bsoncodec.Registry, val reflect.Value) (*structDescription, error) {
	t := val.Type()
	protoMsg := val.Interface().(proto.Message)
	protoDesc := protoMsg.ProtoReflect()
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

	var fields []fieldDescription
	for i := 0; i < numFields; i++ {
		sf := fieldsDesc.Get(i)
		sfv := protoDesc.Get(sf)
		sfValue := reflect.ValueOf(sfv.Interface())
		sfType := sfValue.Type()

		encoder, err := r.LookupEncoder(sfType)
		if err != nil {
			encoder = nil
		}

		decoder, err := r.LookupDecoder(sfType)
		if err != nil {
			decoder = nil
		}

		description := fieldDescription{
			fieldName: string(sf.Name()),
			idx:       i,
			encoder:   encoder,
			decoder:   decoder,
		}

		description.name = sf.JSONName()

		if sf.ContainingOneof() != nil {
			sd.inline = true
			switch sfType.Kind() {
			case reflect.Ptr, reflect.Interface:
				sfType = sfType.Elem()
				if sfType.Kind() != reflect.Struct {
					return nil, fmt.Errorf("(struct %s) inline fields must be a struct or a struct pointer", t.String())
				}

				fallthrough
			case reflect.Struct:
				inlinesf, err := sc.describeStruct(r, sfValue)
				if err != nil {
					return nil, err
				}

				for _, fd := range inlinesf.fl {
					if fd.inline == nil {
						fd.inline = []int{i, fd.idx}
					} else {
						fd.inline = append([]int{i}, fd.inline...)
					}

					fields = append(fields, fd)
				}
			default:
				return nil, fmt.Errorf("(struct %s) inline fields must be a struct or a struct pointer", t.String())
			}

			continue
		}

		fields = append(fields, description)
	}

	// Sort fieldDescriptions by name and use dominance rules to determine which should be added for each name
	sort.Slice(fields, func(i, j int) bool {
		x := fields
		// sort field by name, breaking ties with depth, then
		// breaking ties with index sequence.
		if x[i].name != x[j].name {
			return x[i].name < x[j].name
		}

		if len(x[i].inline) != len(x[j].inline) {
			return len(x[i].inline) < len(x[j].inline)
		}

		return byIndex(x).Less(i, j)
	})

	for advance, i := 0, 0; i < len(fields); i += advance {
		// One iteration per name.
		// Find the sequence of fields with the name of this first field.
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			fj := fields[i+advance]
			if fj.name != name {
				break
			}
		}

		if advance == 1 { // Only one field with this name
			sd.fl = append(sd.fl, fi)
			sd.fm[name] = fi
			continue
		}

		dominant, ok := dominantField(fields[i : i+advance])
		if !ok {
			return nil, fmt.Errorf("struct %s has duplicated key %s", t.String(), name)
		}

		sd.fl = append(sd.fl, dominant)
		sd.fm[name] = dominant
	}

	sort.Sort(byIndex(sd.fl))

	sc.l.Lock()
	sc.cache[t] = sd
	sc.l.Unlock()

	return sd, nil
}

// dominantField looks through the fields, all of which are known to
// have the same name, to find the single field that dominates the
// others using Go's inlining rules. If there are multiple top-level
// fields, the boolean will be false: This condition is an error in Go
// and we skip all the fields.
func dominantField(fields []fieldDescription) (fieldDescription, bool) {
	// The fields are sorted in increasing index-length order, then by presence of tag.
	// That means that the first field is the dominant one. We need only check
	// for error cases: two fields at top level.
	if len(fields) > 1 &&
		len(fields[0].inline) == len(fields[1].inline) {
		return fieldDescription{}, false
	}
	return fields[0], true
}

func fieldByIndexErr(v reflect.Value, index []int) (result reflect.Value, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			switch r := recovered.(type) {
			case string:
				err = fmt.Errorf("%s", r)
			case error:
				err = r
			}
		}
	}()

	result = v.FieldByIndex(index)

	return
}

func getInlineField(val reflect.Value, index []int) (reflect.Value, error) {
	field, err := fieldByIndexErr(val, index)
	if err == nil {
		return field, nil
	}

	// if parent of this element doesn't exist, fix its parent
	inlineParent := index[:len(index)-1]
	var fParent reflect.Value
	if fParent, err = fieldByIndexErr(val, inlineParent); err != nil {
		fParent, err = getInlineField(val, inlineParent)
		if err != nil {
			return fParent, err
		}
	}

	fParent.Set(reflect.New(fParent.Type().Elem()))

	return fieldByIndexErr(val, index)
}

// DeepZero returns recursive zero object
func deepZero(st reflect.Type) (result reflect.Value) {
	result = reflect.Indirect(reflect.New(st))

	if result.Kind() == reflect.Struct {
		for i := 0; i < result.NumField(); i++ {
			if f := result.Field(i); f.Kind() == reflect.Ptr {
				if f.CanInterface() {
					if ft := reflect.TypeOf(f.Interface()); ft.Elem().Kind() == reflect.Struct {
						result.Field(i).Set(recursivePointerTo(deepZero(ft.Elem())))
					}
				}
			}
		}
	}

	return
}

// recursivePointerTo calls reflect.New(v.Type) but recursively for its fields inside
func recursivePointerTo(v reflect.Value) reflect.Value {
	v = reflect.Indirect(v)
	result := reflect.New(v.Type())
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			if f := v.Field(i); f.Kind() == reflect.Ptr {
				if f.Elem().Kind() == reflect.Struct {
					result.Elem().Field(i).Set(recursivePointerTo(f))
				}
			}
		}
	}

	return result
}

// isImplementationNil returns if val is a nil pointer and inter is implemented on a concrete type
func isImplementationNil(val reflect.Value, inter reflect.Type) bool {
	vt := val.Type()
	for vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	return vt.Implements(inter) && val.Kind() == reflect.Ptr && val.IsNil()
}

func lookupElementEncoder(ec bsoncodec.EncodeContext, origEncoder bsoncodec.ValueEncoder, currVal reflect.Value) (bsoncodec.ValueEncoder, reflect.Value, error) {
	if origEncoder != nil || (currVal.Kind() != reflect.Interface) {
		return origEncoder, currVal, nil
	}

	currVal = currVal.Elem()
	if !currVal.IsValid() {
		return nil, currVal, fmt.Errorf("cannot encode invalid element")
	}

	currEncoder, err := ec.LookupEncoder(currVal.Type())

	return currEncoder, currVal, err
}
