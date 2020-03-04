package http

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jexia/maestro/codec/json"
	"github.com/jexia/maestro/flow"
	"github.com/jexia/maestro/protocol"
	"github.com/jexia/maestro/refs"
	"github.com/jexia/maestro/schema"
	"github.com/jexia/maestro/specs"
	"github.com/jexia/maestro/specs/types"
)

type MockService struct {
	name    string
	methods []schema.Method
	options schema.Options
}

func (service *MockService) GetName() string {
	return service.name
}

func (service *MockService) GetMethod(name string) schema.Method {
	for _, method := range service.methods {
		if method.GetName() == name {
			return method
		}
	}

	return nil
}

func (service *MockService) GetMethods() []schema.Method {
	return service.methods
}

func (service *MockService) GetOptions() schema.Options {
	return service.options
}

type MockMethod struct {
	name    string
	options schema.Options
	input   schema.Object
	output  schema.Object
}

func (method *MockMethod) GetName() string {
	return method.name
}

func (method *MockMethod) GetInput() schema.Object {
	return method.input
}

func (method *MockMethod) GetOutput() schema.Object {
	return method.output
}

func (method *MockMethod) GetOptions() schema.Options {
	return method.options
}

type MockResponseWriter struct {
	header protocol.Header
	writer io.Writer
}

func (rw *MockResponseWriter) Header() protocol.Header {
	return rw.header
}

func (rw *MockResponseWriter) Write(bb []byte) (int, error) {
	return rw.writer.Write(bb)
}

func (rw *MockResponseWriter) WriteHeader(int) {}

func AvailablePort(t *testing.T) int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func TestCaller(t *testing.T) {
	message := "hello world"
	specs := &specs.ParameterMap{
		Properties: map[string]*specs.Property{
			"message": &specs.Property{
				Name: "message",
				Path: "message",
				Type: types.TypeString,
			},
		},
	}

	cons := &json.Constructor{}
	codec, err := cons.New("input", specs)
	if err != nil {
		t.Fatal(err)
	}

	refs := refs.NewStore(1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"` + message + `"}`))
	}))

	defer server.Close()

	ctx := context.Background()
	req := protocol.Request{
		Context: ctx,
	}

	service := &MockService{
		methods: []schema.Method{
			&MockMethod{
				options: schema.Options{
					EndpointOption: "/",
					MethodOption:   "GET",
				},
			},
		},
	}
	constructor := &Caller{}
	caller, err := constructor.New(server.URL, service, nil)
	if err != nil {
		t.Fatal(err)
	}

	r, w := io.Pipe()
	rw := &MockResponseWriter{
		header: protocol.Header{},
		writer: w,
	}

	go func() {
		caller.Call(rw, &req, refs)
		w.Close()
	}()

	err = codec.Unmarshal(r, refs)
	if err != nil {
		t.Fatal(err)
	}

	ref := refs.Load("input", "message")
	if ref == nil {
		t.Fatal("input:message reference not set")
	}

	result, is := ref.Value.(string)
	if !is {
		t.Fatal("input:message reference is not a string")
	}

	if result != message {
		t.Fatalf("unexpected input:message %s, expected %s", result, message)
	}
}

func TestListener(t *testing.T) {
	called := 0
	port := AvailablePort(t)
	addr := fmt.Sprintf(":%d", port)
	listener, err := NewListener(addr, nil)
	if err != nil {
		t.Fatal(err)
	}

	defer listener.Close()

	nodes := flow.Nodes{
		{
			Name:     "first",
			Previous: flow.Nodes{},
			Call: func(ctx context.Context, refs *refs.Store) error {
				called++
				return nil
			},
			Next: flow.Nodes{},
		},
	}

	endpoints := []*protocol.Endpoint{
		{
			Flow: flow.NewManager("test", nodes),
			Options: specs.Options{
				"endpoint": "/",
				"method":   "GET",
			},
		},
	}

	listener.Handle(endpoints)
	go listener.Serve()

	// Some CI pipelines take a little while before the listener is active
	time.Sleep(100 * time.Millisecond)

	endpoint := fmt.Sprintf("http://127.0.0.1:%d/", port)
	http.Get(endpoint)

	if called != 1 {
		t.Errorf("unexpected called %d, expected %d", called, len(nodes))
	}
}