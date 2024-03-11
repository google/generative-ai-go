package genai

import (
	"errors"
	"fmt"
	"reflect"
)

// FunctionSchema returns a Schema for a Go function.
// Not all functions can be represented as Schemas.
// At present, variadic functions are not supported, and parameters
// must be of builtin, pointer, slice or array type.
//
// Parameter names are not available to the program. They can be supplied
// as arguments. If omitted, the names "p0", "p1", ... are used.
func FunctionSchema(function any, paramNames ...string) (*Schema, error) {
	t := reflect.TypeOf(function)
	if t == nil || t.Kind() != reflect.Func {
		return nil, fmt.Errorf("value of type %T is not a function", function)
	}
	if t.IsVariadic() {
		return nil, errors.New("variadic functions not supported")
	}
	params := map[string]*Schema{}
	var req []string
	for i := 0; i < t.NumIn(); i++ {
		var name string
		if i < len(paramNames) {
			name = paramNames[i]
		} else {
			name = fmt.Sprintf("p%d", i)
		}
		s, err := typeSchema(t.In(i))
		if err != nil {
			return nil, fmt.Errorf("param %s: %w", name, err)
		}
		params[name] = s
		// All parameters are required.
		req = append(req, name)
	}
	return &Schema{
		Type:       TypeObject,
		Properties: params,
		Required:   req,
	}, nil

}

func typeSchema(t reflect.Type) (_ *Schema, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %w", t, err)
		}
	}()
	switch t.Kind() {
	case reflect.Bool:
		return &Schema{Type: TypeBoolean}, nil
	case reflect.String:
		return &Schema{Type: TypeString}, nil
	case reflect.Int, reflect.Int64, reflect.Uint32:
		return &Schema{Type: TypeInteger, Format: "int64"}, nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint8, reflect.Uint16:
		return &Schema{Type: TypeInteger, Format: "int32"}, nil
	case reflect.Float32:
		return &Schema{Type: TypeNumber, Format: "float"}, nil
	case reflect.Float64, reflect.Uint, reflect.Uint64, reflect.Uintptr:
		return &Schema{Type: TypeNumber, Format: "double"}, nil
	case reflect.Slice, reflect.Array:
		elemSchema, err := typeSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		return &Schema{Type: TypeArray, Items: elemSchema}, nil
	case reflect.Pointer:
		// Treat a *T as a nullable T.
		s, err := typeSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		s.Nullable = true
		return s, nil
	default:
		return nil, errors.New("not supported")
	}
}
