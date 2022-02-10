package internal

import (
	"errors"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var ErrDirectSet = errors.New("field should be accessed directly")

func ReflectIntoStd(sfv protoreflect.Value, useEnumNumbers bool) reflect.Type {
	var sfType reflect.Type

	switch v := sfv.Interface().(type) {
	case protoreflect.EnumNumber:
		if useEnumNumbers {
			return reflect.ValueOf(v).Type()
		}

		sfType = reflect.ValueOf("").Type()
	case protoreflect.Message:
		sfType = reflect.ValueOf(v.Interface()).Type()
	default:
		sfType = reflect.ValueOf(v).Type()
	}

	return sfType
}

func ReflectToProtoValue(
	field reflect.Value,
) (protoreflect.Value, error) {
	switch cv := field.Interface().(type) {
	case protoreflect.ProtoMessage:
		return protoreflect.ValueOf(cv.ProtoReflect()), nil
	default:
		return protoreflect.ValueOf(field.Interface()), nil
	}
}

func GetRWValue(
	rv protoreflect.Value,
	protoMsg protoreflect.ProtoMessage,
	name string,
) (reflect.Value, error) {
	switch v := rv.Interface().(type) {
	case bool:
		return reflect.ValueOf(&v).Elem(), nil
	case int32:
		return reflect.ValueOf(&v).Elem(), nil
	case uint32:
		return reflect.ValueOf(&v).Elem(), nil
	case int64:
		return reflect.ValueOf(&v).Elem(), nil
	case uint64:
		return reflect.ValueOf(&v).Elem(), nil
	case float64:
		return reflect.ValueOf(&v).Elem(), nil
	case float32:
		return reflect.ValueOf(&v).Elem(), nil
	case string:
		return reflect.ValueOf(&v).Elem(), nil
	case []byte:
		return reflect.ValueOf(&v).Elem(), nil
	case protoreflect.Message:
		return reflect.ValueOf(v.Interface()), nil
	case protoreflect.List, protoreflect.Map, protoreflect.EnumNumber:
		// for lists and maps use origin struct to access values since
		// reflected representation loose element type information for
		// enums.
		return reflect.ValueOf(protoMsg).Elem().FieldByName(name), nil
	default:
		return reflect.ValueOf(v), nil
	}
}

func PrepareInterfaceForEncode(where string, val reflect.Value, t reflect.Type) (reflect.Value, error) {
	// Either val or a pointer to val must implement ValueMarshaler
	switch {
	case !val.IsValid():
		return reflect.Value{}, bsoncodec.ValueEncoderError{Name: where, Types: []reflect.Type{t}, Received: val}
	case val.Type().Implements(t):
	case reflect.PtrTo(val.Type()).Implements(t) && val.CanAddr():
		val = val.Addr()
	default:
		return reflect.Value{}, bsoncodec.ValueEncoderError{Name: where, Types: []reflect.Type{t}, Received: val}
	}

	return val, nil
}

func PrepareInterfaceForDecode(where string, val reflect.Value, t reflect.Type) (reflect.Value, error) {
	if !val.IsValid() || (!val.Type().Implements(t) && !reflect.PtrTo(val.Type()).Implements(t)) {
		return reflect.Value{}, bsoncodec.ValueDecoderError{Name: where, Types: []reflect.Type{t}, Received: val}
	}

	if val.Kind() == reflect.Ptr && val.IsNil() {
		if !val.CanSet() {
			return reflect.Value{}, bsoncodec.ValueDecoderError{Name: where, Types: []reflect.Type{t}, Received: val}
		}

		val.Set(reflect.New(val.Type().Elem()))
	}

	if !val.Type().Implements(t) {
		if !val.CanAddr() {
			return reflect.Value{}, bsoncodec.ValueDecoderError{Name: where, Types: []reflect.Type{t}, Received: val}
		}

		val = val.Addr() // If they type doesn't implement the interface, a pointer to it must.
	}

	return val, nil
}

func PrepareValueForEncode(where string, val reflect.Value, t reflect.Type) (reflect.Value, error) {
	// Either val or a pointer to val must implement ValueMarshaler
	switch {
	case !val.IsValid():
		return reflect.Value{}, bsoncodec.ValueEncoderError{Name: where, Types: []reflect.Type{t}, Received: val}
	case val.Type().AssignableTo(t):
	case reflect.PtrTo(val.Type()).AssignableTo(t) && val.CanAddr():
		val = val.Addr()
	default:
		return reflect.Value{}, bsoncodec.ValueEncoderError{Name: where, Types: []reflect.Type{t}, Received: val}
	}

	return val, nil
}

func PrepareValueForDecode(where string, val reflect.Value, t reflect.Type) (reflect.Value, error) {
	if !val.IsValid() || (!val.Type().AssignableTo(t) && !reflect.PtrTo(val.Type()).AssignableTo(t)) {
		return reflect.Value{}, bsoncodec.ValueDecoderError{Name: where, Types: []reflect.Type{t}, Received: val}
	}

	if val.Kind() == reflect.Ptr && val.IsNil() {
		if !val.CanSet() {
			return reflect.Value{}, bsoncodec.ValueDecoderError{Name: where, Types: []reflect.Type{t}, Received: val}
		}

		val.Set(reflect.New(val.Type().Elem()))
	}

	if !val.Type().AssignableTo(t) {
		if !val.CanAddr() {
			return reflect.Value{}, bsoncodec.ValueDecoderError{Name: where, Types: []reflect.Type{t}, Received: val}
		}

		val = val.Addr() // If they type doesn't implement the interface, a pointer to it must.
	}

	return val, nil
}
