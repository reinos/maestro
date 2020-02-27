package json

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/francoispqt/gojay"
	"github.com/jexia/maestro/codec"
	"github.com/jexia/maestro/refs"
	"github.com/jexia/maestro/specs"
	"github.com/jexia/maestro/specs/trace"
)

// Constructor is capable of constructing new codec managers for the given resource and specs
type Constructor struct {
}

func (constructor *Constructor) Name() string {
	return "json"
}

func (constructor *Constructor) New(resource string, specs specs.Object) (codec.Manager, error) {
	if specs == nil {
		return nil, trace.New(trace.WithMessage("no object specs defined"))
	}

	return &Manager{
		resource: resource,
		specs:    specs,
	}, nil
}

// Manager manages a specs object and allows to encode/decode messages
type Manager struct {
	resource string
	specs    specs.Object
	keys     int
}

// Marshal marshals the given reference store into a JSON message.
// This method is called during runtime to encode a new message with the values stored inside the given reference store
func (manager *Manager) Marshal(refs *refs.Store) (io.Reader, error) {
	object := NewObject(manager.resource, manager.specs, refs)
	bb, err := gojay.MarshalJSONObject(object)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(bb), nil
}

// Unmarshal unmarshals the given JSON io reader into the given reference store.
// This method is called during runtime to decode a new message and store it inside the given reference store
func (manager *Manager) Unmarshal(reader io.Reader, refs *refs.Store) error {
	bb, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	object := NewObject(manager.resource, manager.specs, refs)
	err = gojay.UnmarshalJSONObject(bb, object)
	if err != nil {
		return err
	}

	return nil
}
