package proto

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jexia/maestro"
	"github.com/jexia/maestro/definitions/hcl"
	"github.com/jexia/maestro/refs"
	"github.com/jexia/maestro/schema/protoc"
	"github.com/jexia/maestro/specs"
	"github.com/jhump/protoreflect/dynamic"
)

func NewMock() (*specs.Manifest, error) {
	collection, err := protoc.Collect([]string{"./tests"}, "./tests/*.proto")
	if err != nil {
		return nil, err
	}

	client, err := maestro.New(maestro.WithDefinitions(hcl.DefinitionResolver("./tests/*.hcl")), maestro.WithSchema(collection))
	if err != nil {
		return nil, err
	}

	return client.Manifest, nil
}

func FindFlow(manifest *specs.Manifest, name string) *specs.Flow {
	for _, flow := range manifest.Flows {
		if flow.GetName() == name {
			return flow
		}
	}

	return nil
}

func FindNode(flow *specs.Flow, name string) *specs.Node {
	for _, node := range flow.GetNodes() {
		if node.GetName() == name {
			return node
		}
	}

	return nil
}

func ValidateStore(t *testing.T, resource string, origin string, input map[string]interface{}, store *refs.Store) {
	for key, value := range input {
		path := specs.JoinPath(origin, key)
		nested, is := value.(map[string]interface{})
		if is {
			ValidateStore(t, resource, path, nested, store)
			continue
		}

		repeated, is := value.([]map[string]interface{})
		if is {
			repeating := store.Load(resource, path)
			if repeating == nil {
				t.Fatalf("repeating message does not exist in store '%s.%s'", resource, path)
			}

			for index, store := range repeating.Repeated {
				ValidateStore(t, resource, path, repeated[index], store)
			}
			continue
		}

		ref := store.Load(resource, path)
		if ref == nil {
			t.Fatalf("resource not found %s", path)
		}

		if ref.Value != value {
			t.Fatalf("unexpected value at %s '%+v', expected '%+v'", path, ref.Value, value)
		}
	}
}

func BenchmarkSimpleMarshal(b *testing.B) {
	input := map[string]interface{}{
		"message": "message",
	}

	refs := refs.NewStore(len(input))
	refs.StoreValues("input", "", input)

	manifest, err := NewMock()
	if err != nil {
		b.Fatal(err)
	}

	flow := FindFlow(manifest, "simple")
	specs := FindNode(flow, "first").Call.GetRequest()

	constructor := NewConstructor()
	manager, err := constructor.New("input", specs)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader, err := manager.Marshal(refs)
		if err != nil {
			b.Fatal(err)
		}

		ioutil.ReadAll(reader)
	}
}

func BenchmarkNestedMarshal(b *testing.B) {
	input := map[string]interface{}{
		"nested": map[string]interface{}{
			"value": "message",
		},
	}

	refs := refs.NewStore(len(input))
	refs.StoreValues("input", "", input)

	manifest, err := NewMock()
	if err != nil {
		b.Fatal(err)
	}

	flow := FindFlow(manifest, "simple")
	specs := FindNode(flow, "first").Call.GetRequest()

	constructor := NewConstructor()
	manager, err := constructor.New("input", specs)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader, err := manager.Marshal(refs)
		if err != nil {
			b.Fatal(err)
		}

		ioutil.ReadAll(reader)
	}
}

func BenchmarkRepeatedMarshal(b *testing.B) {
	input := map[string]interface{}{
		"repeating": []map[string]interface{}{
			{
				"value": "message",
			},
		},
	}

	refs := refs.NewStore(len(input))
	refs.StoreValues("input", "", input)

	manifest, err := NewMock()
	if err != nil {
		b.Fatal(err)
	}

	flow := FindFlow(manifest, "simple")
	specs := FindNode(flow, "first").Call.GetRequest()

	constructor := NewConstructor()
	manager, err := constructor.New("input", specs)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader, err := manager.Marshal(refs)
		if err != nil {
			b.Fatal(err)
		}

		ioutil.ReadAll(reader)
	}
}

func BenchmarkSimpleUnmarshal(b *testing.B) {
	input := map[string]interface{}{
		"message": "message",
	}

	jsonBB, err := json.Marshal(input)
	if err != nil {
		b.Fatal(err)
	}

	refs := refs.NewStore(len(input))
	manifest, err := NewMock()
	if err != nil {
		b.Fatal(err)
	}

	flow := FindFlow(manifest, "simple")
	specs := FindNode(flow, "first").Call.GetRequest()

	desc, err := NewMessage(specs.Property.Name, specs.Property.Nested)
	if err != nil {
		b.Fatal(err)
	}

	msg := dynamic.NewMessage(desc)
	err = msg.UnmarshalJSON(jsonBB)
	if err != nil {
		b.Fatal(err)
	}

	bb, err := msg.Marshal()
	if err != nil {
		b.Fatal(err)
	}

	constructor := NewConstructor()
	manager, err := constructor.New("input", specs)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := manager.Unmarshal(bytes.NewBuffer(bb), refs)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNestedUnmarshal(b *testing.B) {
	input := map[string]interface{}{
		"nested": map[string]interface{}{
			"value": "message",
		},
	}

	jsonBB, err := json.Marshal(input)
	if err != nil {
		b.Fatal(err)
	}

	refs := refs.NewStore(len(input))
	manifest, err := NewMock()
	if err != nil {
		b.Fatal(err)
	}

	flow := FindFlow(manifest, "nested")
	specs := FindNode(flow, "first").Call.GetRequest()

	desc, err := NewMessage(specs.Property.Name, specs.Property.Nested)
	if err != nil {
		b.Fatal(err)
	}

	msg := dynamic.NewMessage(desc)
	err = msg.UnmarshalJSON(jsonBB)
	if err != nil {
		b.Fatal(err)
	}

	bb, err := msg.Marshal()
	if err != nil {
		b.Fatal(err)
	}

	constructor := NewConstructor()
	manager, err := constructor.New("input", specs)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := manager.Unmarshal(bytes.NewBuffer(bb), refs)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepeatedUnmarshal(b *testing.B) {
	input := map[string]interface{}{
		"repeating": []map[string]interface{}{
			{
				"value": "message",
			},
		},
	}

	jsonBB, err := json.Marshal(input)
	if err != nil {
		b.Fatal(err)
	}

	refs := refs.NewStore(len(input))
	manifest, err := NewMock()
	if err != nil {
		b.Fatal(err)
	}

	flow := FindFlow(manifest, "repeated")
	specs := FindNode(flow, "first").Call.GetRequest()

	desc, err := NewMessage(specs.Property.Name, specs.Property.Nested)
	if err != nil {
		b.Fatal(err)
	}

	msg := dynamic.NewMessage(desc)
	err = msg.UnmarshalJSON(jsonBB)
	if err != nil {
		b.Fatal(err)
	}

	bb, err := msg.Marshal()
	if err != nil {
		b.Fatal(err)
	}

	constructor := NewConstructor()
	manager, err := constructor.New("input", specs)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := manager.Unmarshal(bytes.NewBuffer(bb), refs)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestMarshal(t *testing.T) {
	manifest, err := NewMock()
	if err != nil {
		t.Fatal(err)
	}

	flow := FindFlow(manifest, "complete")
	specs := FindNode(flow, "first").Call.GetRequest()
	desc, err := NewMessage("marshal", specs.Property.Nested)
	if err != nil {
		t.Fatal(err)
	}

	response := dynamic.NewMessage(desc)

	constructor := NewConstructor()
	manager, err := constructor.New("input", specs)
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]map[string]interface{}{
		"simple": {
			"message": "hello world",
			"nested":  map[string]interface{}{},
		},
		"nesting": {
			"nested": map[string]interface{}{
				"value": "nested value",
			},
		},
		"complex": {
			"message": "hello world",
			"nested": map[string]interface{}{
				"value": "nested value",
			},
			"repeating": []map[string]interface{}{
				{
					"value": "repeating value",
				},
			},
		},
	}

	for key, input := range tests {
		t.Run(key, func(t *testing.T) {
			inputAsJSON, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}

			store := refs.NewStore(3)
			store.StoreValues("input", "", input)

			reader, err := manager.Marshal(store)
			if err != nil {
				t.Fatal(err)
			}

			bb, err := ioutil.ReadAll(reader)
			if err != nil {
				t.Fatal(err)
			}

			err = response.Unmarshal(bb)
			if err != nil {
				t.Fatal(err)
			}

			responseAsJSON, err := response.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}

			result := map[string]interface{}{}
			err = json.Unmarshal(responseAsJSON, &result)
			if err != nil {
				t.Fatal(err)
			}

			expected := map[string]interface{}{}
			err = json.Unmarshal(inputAsJSON, &expected)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(expected, result) {
				t.Errorf("unexpected response %s, expected %s", string(responseAsJSON), string(inputAsJSON))
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	manifest, err := NewMock()
	if err != nil {
		t.Fatal(err)
	}

	flow := FindFlow(manifest, "complete")
	specs := FindNode(flow, "first").Call.GetRequest()

	tests := map[string]map[string]interface{}{
		"simple": {
			"message": "hello world",
			"nested":  map[string]interface{}{},
		},
		"nested": {
			"nested": map[string]interface{}{
				"value": "nested value",
			},
		},
		"complex": {
			"message": "hello world",
			"nested": map[string]interface{}{
				"value": "nested value",
			},
			"repeating": []map[string]interface{}{
				{
					"value": "repeating value",
				},
			},
		},
	}

	for key, input := range tests {
		t.Run(key, func(t *testing.T) {
			inputAsJSON, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}

			desc, err := NewMessage("input", specs.Property.Nested)
			if err != nil {
				t.Fatal(err)
			}

			inputAsProto := dynamic.NewMessage(desc)
			err = inputAsProto.UnmarshalJSON(inputAsJSON)
			if err != nil {
				t.Fatal(err)
			}

			bb, err := inputAsProto.Marshal()
			if err != nil {
				t.Fatal(err)
			}

			store := refs.NewStore(len(input))

			constructor := NewConstructor()
			manager, err := constructor.New("input", specs)
			if err != nil {
				t.Fatal(err)
			}

			err = manager.Unmarshal(bytes.NewReader(bb), store)
			if err != nil {
				t.Fatal(err)
			}

			t.Log(store)

			ValidateStore(t, "input", "", input, store)
		})
	}
}
