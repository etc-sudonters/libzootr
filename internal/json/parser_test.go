package json

import "testing"

func TestCanDiscard(t *testing.T) {
	input := []byte(`{"array": [1,2,3,4,5], "nested": {"property": 1}, "another": {}}`)
	parser := NewParser(scanner(input))
	obj, err := parser.ReadObject()
	fatalOnErr(t, err)
	obj.DiscardRemaining()
	if parser.Next() {
		t.Logf("found next %q", string(parser.Current().Body))
		t.Fatalf("expected to discard entire object")
	}
}

func TestCanParseEmptyObj(t *testing.T) {
	input := []byte(`{}`)
	parser := NewParser(scanner(input))

	obj, err := parser.ReadObject()
	if err != nil {
		t.Fatalf("expected object: %s", err)
	}

	if err := obj.ReadEnd(); err != nil {
		t.Fatalf("expected object end: %s", err)
	}
}

func TestCanParseEmptyArr(t *testing.T) {
	input := []byte(`[]`)
	parser := NewParser(scanner(input))

	obj, err := parser.ReadArray()
	if err != nil {
		t.Fatalf("expected object: %s", err)
	}

	if err := obj.ReadEnd(); err != nil {
		t.Fatalf("expected object end: %s", err)
	}
}

func TestCanParseComplicated(t *testing.T) {
	input := []byte(`
    {
        "property" #comment
        : "words words words",
        "nested": [1,2, //comment
        3
        ]
    }
`)

	type test struct {
		property string
		values   []int
	}

	parser := NewParser(scanner(input))
	obj, err := parser.ReadObject()
	fatalOnErr(t, err)
	building := test{}

	for obj.More() {
		property, err := obj.ReadPropertyName()
		fatalOnErr(t, err)
		switch property {
		case "property":
			value, err := obj.ReadString()
			fatalOnErr(t, err)
			building.property = value
		case "nested":
			arr, err := obj.ReadArray()
			fatalOnErr(t, err)
			for arr.More() {
				value, err := arr.ReadInt()
				fatalOnErr(t, err)
				building.values = append(building.values, value)
			}
			fatalOnErr(t, arr.ReadEnd())
		}
	}
	fatalOnErr(t, obj.ReadEnd())

	if building.property != "words words words" {
		t.Fail()
		t.Logf("expected to set %q but set %q", "words words words", building.property)
	}

	if len(building.values) != 3 {
		t.Fail()
		t.Logf("expected to read %d integers, read %d", 3, len(building.values))
	}

	for i := range min(3, len(building.values)) {
		if i+1 != building.values[i] {
			t.Fail()
			t.Logf("expected(%d) to be %d but was %d", i, i+1, building.values[i])
		}
	}

}

func fatalOnErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
