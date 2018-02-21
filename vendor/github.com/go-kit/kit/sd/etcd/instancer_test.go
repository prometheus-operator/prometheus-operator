package etcd

import (
	"errors"
	"testing"

	stdetcd "github.com/coreos/etcd/client"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
)

var _ sd.Instancer = (*Instancer)(nil) // API check

var (
	node = &stdetcd.Node{
		Key: "/foo",
		Nodes: []*stdetcd.Node{
			{Key: "/foo/1", Value: "1:1"},
			{Key: "/foo/2", Value: "1:2"},
		},
	}
	fakeResponse = &stdetcd.Response{
		Node: node,
	}
)

var _ sd.Instancer = &Instancer{} // API check

func TestInstancer(t *testing.T) {
	client := &fakeClient{
		responses: map[string]*stdetcd.Response{"/foo": fakeResponse},
	}

	s, err := NewInstancer(client, "/foo", log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	if state := s.cache.State(); state.Err != nil {
		t.Fatal(state.Err)
	}
}

type fakeClient struct {
	responses map[string]*stdetcd.Response
}

func (c *fakeClient) GetEntries(prefix string) ([]string, error) {
	response, ok := c.responses[prefix]
	if !ok {
		return nil, errors.New("key not exist")
	}

	entries := make([]string, len(response.Node.Nodes))
	for i, node := range response.Node.Nodes {
		entries[i] = node.Value
	}
	return entries, nil
}

func (c *fakeClient) WatchPrefix(prefix string, ch chan struct{}) {}

func (c *fakeClient) Register(Service) error {
	return nil
}
func (c *fakeClient) Deregister(Service) error {
	return nil
}
