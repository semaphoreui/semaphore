package util

import (
    "encoding/json"
    "reflect"
    "testing"
)


func TestAssignMapToStruct_SlicesAndConversions(t *testing.T) {
    type Item struct {
        K string `json:"k"`
        V int    `json:"v"`
    }

    type Sample struct {
        Names    []string        `json:"names"`
        Numbers  []int           `json:"numbers"`
        Objects  []Item          `json:"objects"`
        Enabled  bool            `json:"enabled"`
        Count    int             `json:"count"`
        Settings map[string]Item `json:"settings"`
    }

    t.Run("primitive slice from json string and fallback single string", func(t *testing.T) {
        var s Sample
        m := map[string]any{
            "names": "[\"a\",\"b\"]",
        }
        if err := AssignMapToStruct(m, &s); err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !reflect.DeepEqual(s.Names, []string{"a", "b"}) {
            t.Fatalf("names mismatch: %+v", s.Names)
        }

        // fallback: non-JSON string becomes single element when elem type is string
        s = Sample{}
        m = map[string]any{"names": "hello"}
        if err := AssignMapToStruct(m, &s); err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !reflect.DeepEqual(s.Names, []string{"hello"}) {
            t.Fatalf("names fallback mismatch: %+v", s.Names)
        }
    })

    t.Run("int slice with mixed string/int and coercion", func(t *testing.T) {
        var s Sample
        // input as a real slice ([]any) with mixed types
        src := []any{"1", 2, "3"}
        m := map[string]any{"numbers": src}
        if err := AssignMapToStruct(m, &s); err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !reflect.DeepEqual(s.Numbers, []int{1, 2, 3}) {
            t.Fatalf("numbers mismatch: %+v", s.Numbers)
        }

        // input as JSON string
        s = Sample{}
        jsonStr := "[\"4\",5,\"6\"]"
        m = map[string]any{"numbers": jsonStr}
        if err := AssignMapToStruct(m, &s); err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !reflect.DeepEqual(s.Numbers, []int{4, 5, 6}) {
            t.Fatalf("numbers from json mismatch: %+v", s.Numbers)
        }
    })

    t.Run("slice of structs from []map and JSON string of maps", func(t *testing.T) {
        var s Sample
        objs := []any{
            map[string]any{"k": "a", "v": 1},
            map[string]any{"k": "b", "v": "2"}, // v as string, should coerce to int
        }
        m := map[string]any{"objects": objs}
        if err := AssignMapToStruct(m, &s); err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if len(s.Objects) != 2 || s.Objects[0].K != "a" || s.Objects[0].V != 1 || s.Objects[1].K != "b" || s.Objects[1].V != 2 {
            t.Fatalf("objects mismatch: %+v", s.Objects)
        }

        // JSON string input
        s = Sample{}
        arr := []map[string]any{{"k": "x", "v": 7}, {"k": "y", "v": 8}}
        b, _ := json.Marshal(arr)
        m = map[string]any{"objects": string(b)}
        if err := AssignMapToStruct(m, &s); err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if len(s.Objects) != 2 || s.Objects[0].K != "x" || s.Objects[0].V != 7 || s.Objects[1].K != "y" || s.Objects[1].V != 8 {
            t.Fatalf("objects from json mismatch: %+v", s.Objects)
        }
    })

    t.Run("map update preserves existing nested struct fields", func(t *testing.T) {
        type Detail struct{
            Value       string `json:"value"`
            Description string `json:"description"`
        }
        type Holder struct{ Details map[string]Detail `json:"details"` }
        h := Holder{Details: map[string]Detail{
            "interests": {Value: "politics", Description: "Follows current events"},
        }}
        m := map[string]any{"details": map[string]any{
            "interests": map[string]any{"description": "Ho ho ho"},
        }}
        if err := AssignMapToStruct(m, &h); err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if h.Details["interests"].Value != "politics" || h.Details["interests"].Description != "Ho ho ho" {
            t.Fatalf("map preservation failed: %+v", h.Details["interests"])
        }
    })

    t.Run("primitive field conversions string->int and string->bool", func(t *testing.T) {
        type Conv struct {
            I int  `json:"i"`
            B bool `json:"b"`
        }
        var c Conv
        m := map[string]any{"i": "42", "b": "true"}
        if err := AssignMapToStruct(m, &c); err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if c.I != 42 || c.B != true {
            t.Fatalf("conversions mismatch: %+v", c)
        }
    })

    t.Run("error cases: wrong types for struct and slices", func(t *testing.T) {
        // wrong nested struct source
        type Nested struct{ A struct{ X int } `json:"a"` }
        var n Nested
        if err := AssignMapToStruct(map[string]any{"a": 123}, &n); err == nil {
            t.Fatalf("expected error for non-map nested struct input")
        }

        // wrong slice: non-JSON string for []int should error
        type S struct{ N []int `json:"n"` }
        var s S
        if err := AssignMapToStruct(map[string]any{"n": "not-json"}, &s); err == nil {
            t.Fatalf("expected error for non-JSON string to []int")
        }
    })
}

func TestAssignMapToStruct_MapPrimitiveConversions(t *testing.T) {
    type C struct{
        M map[string]int `json:"m"`
    }
    var c C
    m := map[string]any{"m": map[string]any{"a": "1", "b": 2}}
    if err := AssignMapToStruct(m, &c); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(c.M) != 2 || c.M["a"] != 1 || c.M["b"] != 2 {
        t.Fatalf("map primitive conversions mismatch: %+v", c.M)
    }
}

func TestAssignMapToStruct_SkipsDbMinusTag(t *testing.T) {
	type Sample struct {
		Name     string `json:"name"`
		Password string `json:"password" db:"-"`
		Age      int    `json:"age"`
		Secret   string `json:"secret" db:"-"`
	}

	t.Run("fields with db:- tag should not be assigned", func(t *testing.T) {
		s := Sample{
			Name:     "original",
			Password: "original_password",
			Age:      25,
			Secret:   "original_secret",
		}

		m := map[string]any{
			"name":     "updated",
			"password": "new_password",
			"age":      30,
			"secret":   "new_secret",
		}

		if err := AssignMapToStruct(m, &s); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Fields without db:"-" should be updated
		if s.Name != "updated" {
			t.Errorf("expected Name to be 'updated', got '%s'", s.Name)
		}
		if s.Age != 30 {
			t.Errorf("expected Age to be 30, got %d", s.Age)
		}

		// Fields with db:"-" should retain original values
		if s.Password != "original_password" {
			t.Errorf("expected Password to remain 'original_password', got '%s'", s.Password)
		}
		if s.Secret != "original_secret" {
			t.Errorf("expected Secret to remain 'original_secret', got '%s'", s.Secret)
		}
	})

	t.Run("nested struct with db:- tag fields", func(t *testing.T) {
		type Inner struct {
			Public  string `json:"public"`
			Private string `json:"private" db:"-"`
		}
		type Outer struct {
			Inner Inner `json:"inner"`
		}

		o := Outer{
			Inner: Inner{
				Public:  "original_public",
				Private: "original_private",
			},
		}

		m := map[string]any{
			"inner": map[string]any{
				"public":  "updated_public",
				"private": "updated_private",
			},
		}

		if err := AssignMapToStruct(m, &o); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if o.Inner.Public != "updated_public" {
			t.Errorf("expected Inner.Public to be 'updated_public', got '%s'", o.Inner.Public)
		}
		if o.Inner.Private != "original_private" {
			t.Errorf("expected Inner.Private to remain 'original_private', got '%s'", o.Inner.Private)
		}
	})
}

func TestSetConfigValue_SliceAndMap(t *testing.T) {
    // This ensures setConfigValue (used by defaults/env) is compatible with slice/map JSON
    type X struct {
        Arr []string
        Mp  map[string]int
        I   int
    }
    var x X
    // slice
    setConfigValue(reflect.ValueOf(&x).Elem().FieldByName("Arr"), "[\"a\",\"b\"]")
    if !reflect.DeepEqual(x.Arr, []string{"a", "b"}) {
        t.Fatalf("setConfigValue slice mismatch: %+v", x.Arr)
    }
    // map
    setConfigValue(reflect.ValueOf(&x).Elem().FieldByName("Mp"), "{\"a\":1}")
    if x.Mp["a"] != 1 {
        t.Fatalf("setConfigValue map mismatch: %+v", x.Mp)
    }
    // primitive
    setConfigValue(reflect.ValueOf(&x).Elem().FieldByName("I"), "123")
    if x.I != 123 {
        t.Fatalf("setConfigValue primitive mismatch: %+v", x.I)
    }
}
