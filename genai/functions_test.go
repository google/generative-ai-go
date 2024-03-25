package genai

import (
	"reflect"
	"testing"
)

var intSchema = &Schema{Type: TypeInteger, Format: "int64"}

func TestTypeSchema(t *testing.T) {
	for _, test := range []struct {
		in   any
		want *Schema
	}{
		{true, &Schema{Type: TypeBoolean}},
		{"", &Schema{Type: TypeString}},
		{1, intSchema},
		{byte(1), &Schema{Type: TypeInteger, Format: "int32"}},
		{1.2, &Schema{Type: TypeNumber, Format: "double"}},
		{float32(1.2), &Schema{Type: TypeNumber, Format: "float"}},
		{new(int), &Schema{Type: TypeInteger, Format: "int64", Nullable: true}},
		{
			[]int{},
			&Schema{Type: TypeArray, Items: intSchema},
		},
	} {
		got, err := typeSchema(reflect.TypeOf(test.in))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%T:\ngot  %+v\nwant %+v", test.in, got, test.want)
		}
	}
}

func TestInferSchema(t *testing.T) {
	f := func(a int, b string, c float64) int { return 0 }
	got, err := inferSchema(f, []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	want := &Schema{
		Type: TypeObject,
		Properties: map[string]*Schema{
			"a":  intSchema,
			"b":  {Type: TypeString},
			"p2": {Type: TypeNumber, Format: "double"},
		},
		Required: []string{"a", "b", "p2"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot  %+v\nwant %+v", got, want)
	}
}

func TestInferSchemaErrors(t *testing.T) {
	for i, f := range []any{
		nil,
		3,                 // not a function
		func(x ...int) {}, // variadic
		func(x any) {},    // unsupported type
	} {
		_, err := inferSchema(f, nil)
		if err == nil {
			t.Errorf("#%d: got nil, want error", i)
		}
	}
}
