package protobson

import (
	"errors"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/reflect/protoreflect"
)

var errDirectSet = errors.New("field should be accessed directly")

func reflectIntoStd(sfv protoreflect.Value, useEnumNumbers bool) reflect.Type {
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

func reflectToProtoValue(
	field reflect.Value,
) (protoreflect.Value, error) {
	switch cv := field.Interface().(type) {
	case protoreflect.ProtoMessage:
		return protoreflect.ValueOf(cv.ProtoReflect()), nil
	default:
		return protoreflect.ValueOf(field.Interface()), nil
	}
}

func (sc *ProtoMessageCodec) getRWValue(
	rv protoreflect.Value,
	protoMsg protoreflect.ProtoMessage,
	name string,
	fd protoreflect.FieldDescriptor,
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

func makeAddresableReflection(rv protoreflect.Value, useEnumNumbers bool) reflect.Value {
	// unwrap value to get it addresable
	switch v := rv.Interface().(type) {
	case bool:
		return reflect.ValueOf(&v).Elem()
	case int32:
		return reflect.ValueOf(&v).Elem()
	case uint32:
		return reflect.ValueOf(&v).Elem()
	case int64:
		return reflect.ValueOf(&v).Elem()
	case uint64:
		return reflect.ValueOf(&v).Elem()
	case float64:
		return reflect.ValueOf(&v).Elem()
	case float32:
		return reflect.ValueOf(&v).Elem()
	case string:
		return reflect.ValueOf(&v).Elem()
	case []byte:
		return reflect.ValueOf(&v).Elem()
	case protoreflect.EnumNumber:
		if useEnumNumbers {
			return reflect.ValueOf(&v).Elem()
		}

		str := ""

		return reflect.ValueOf(&str).Elem()
	case protoreflect.Map:
		return reflect.ValueOf(&v).Elem()
	case protoreflect.List:
		return reflect.ValueOf(&v).Elem()
	case protoreflect.Message:
		return reflect.ValueOf(rv.Message().Interface())
	default:
		panic(fmt.Sprintf("unexpected type: %T", v))
	}
}
