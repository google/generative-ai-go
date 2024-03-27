package genai

import (
	"errors"
	"fmt"
	"reflect"

	pb "cloud.google.com/go/ai/generativelanguage/apiv1beta/generativelanguagepb"
	"github.com/google/generative-ai-go/internal/support"
)

// A Tool is a piece of code that enables the system to interact with
// external systems to perform an action, or set of actions, outside of
// knowledge and scope of the model.
type Tool struct {
	// A list of `FunctionDeclarations` available to the model that can
	// be used for function calling.
	//
	// The model or system does not execute the function. Instead the defined
	// function may be returned as a [FunctionCall]
	// with arguments to the client side for execution. The model may decide to
	// call a subset of these functions by populating
	// [FunctionCall][content.part.function_call] in the response. The next
	// conversation turn may contain a
	// [FunctionResponse][content.part.function_response]
	// with the [content.role] "function" generation context for the next model
	// turn.
	FunctionDeclarations []*FunctionDeclaration
}

func (v *Tool) toProto() *pb.Tool {
	if v == nil {
		return nil
	}
	return &pb.Tool{
		FunctionDeclarations: support.TransformSlice(v.FunctionDeclarations, (*FunctionDeclaration).toProto),
	}
}

func (Tool) fromProto(p *pb.Tool) *Tool {
	if p == nil {
		return nil
	}
	return &Tool{
		FunctionDeclarations: support.TransformSlice(p.FunctionDeclarations, (FunctionDeclaration{}).fromProto),
	}
}

// FunctionDeclaration is structured representation of a function declaration as defined by the
// [OpenAPI 3.03 specification](https://spec.openapis.org/oas/v3.0.3). Included
// in this declaration are the function name and parameters.
// Combine FunctionDeclarations into Tools for use in a [ChatSession].
type FunctionDeclaration struct {
	// Required. The name of the function.
	// Must be a-z, A-Z, 0-9, or contain underscores and dashes, with a maximum
	// length of 63.
	Name string
	// Required. A brief description of the function.
	Description string
	// Optional. Describes the parameters to this function.
	Parameters *Schema
	// If set the Go function to call automatically. Its signature must match
	// the schema. Call [NewCallableFunctionDeclaration] to create a FunctionDeclaration
	// with schema inferred from the function itself.
	Function any
}

func (v *FunctionDeclaration) toProto() *pb.FunctionDeclaration {
	if v == nil {
		return nil
	}
	return &pb.FunctionDeclaration{
		Name:        v.Name,
		Description: v.Description,
		Parameters:  v.Parameters.toProto(),
	}
}

func (FunctionDeclaration) fromProto(p *pb.FunctionDeclaration) *FunctionDeclaration {
	if p == nil {
		return nil
	}
	return &FunctionDeclaration{
		Name:        p.Name,
		Description: p.Description,
		Parameters:  (Schema{}).fromProto(p.Parameters),
	}
}

// NewCallableFunctionDeclaration creates a [FunctionDeclaration] from a Go
// function. When added to a [ChatSession], the function will be called
// automatically when the model requests it.
//
// This function infers the schema ([FunctionDeclaration.Parameters]) from the
// function.  Not all functions can be represented as Schemas.
// At present, variadic functions are not supported, and parameters
// must be of builtin, pointer, slice or array type.
// An error is returned if the schema cannot be inferred.
// It may still be possible to construct a usable schema for the function; if so,
// build a [FunctionDeclaration] by hand, setting its exported fields.
//
// Parameter names are not available to the program. They can be supplied
// as arguments. If omitted, the names "p0", "p1", ... are used.
func NewCallableFunctionDeclaration(name, description string, function any, paramNames ...string) (*FunctionDeclaration, error) {
	schema, err := inferSchema(function, paramNames)
	if err != nil {
		return nil, err
	}
	return &FunctionDeclaration{
		Name:        name,
		Description: description,
		Parameters:  schema,
		Function:    function,
	}, nil
}

func inferSchema(function any, paramNames []string) (*Schema, error) {
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
