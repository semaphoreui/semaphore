package conv

import "testing"

func TestStructToFlatMapArrayField(t *testing.T) {
	type Example struct {
		Numbers [2]int
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("StructToFlatMap panicked: %v", r)
		}
	}()
	_ = StructToFlatMap(Example{Numbers: [2]int{1, 2}})
}

func TestStructToFlatMapPointerStruct(t *testing.T) {
	type Inner struct {
		Name string
	}
	type Outer struct {
		Inner *Inner
	}

	m := StructToFlatMap(Outer{Inner: &Inner{Name: "value"}})
	if v, ok := m["Inner.Name"]; !ok || v != "value" {
		t.Fatalf("expected Inner.Name to be 'value', got %v", m)
	}
}

func TestStructToFlatMapNilPointer(t *testing.T) {
	type Outer struct {
		Inner *int
	}

	m := StructToFlatMap(Outer{})
	if v, ok := m["Inner"]; !ok || v != nil {
		t.Fatalf("expected Inner to be nil, got %v", m)
	}
}
