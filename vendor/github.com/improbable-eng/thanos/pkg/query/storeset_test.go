package query

import (
	"context"
	"math"
	"net"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/improbable-eng/thanos/pkg/store/storepb"
	"github.com/improbable-eng/thanos/pkg/testutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var testGRPCOpts = []grpc.DialOption{
	grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)),
	grpc.WithInsecure(),
}

type testStore struct {
	info storepb.InfoResponse
}

func (s *testStore) Info(ctx context.Context, r *storepb.InfoRequest) (*storepb.InfoResponse, error) {
	return &s.info, nil
}

func (s *testStore) Series(r *storepb.SeriesRequest, srv storepb.Store_SeriesServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}

func (s *testStore) LabelNames(ctx context.Context, r *storepb.LabelNamesRequest) (
	*storepb.LabelNamesResponse, error,
) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *testStore) LabelValues(ctx context.Context, r *storepb.LabelValuesRequest) (
	*storepb.LabelValuesResponse, error,
) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

type testStores struct {
	srvs map[string]*grpc.Server
}

func newTestStores(numStores int) (*testStores, error) {
	st := &testStores{
		srvs: map[string]*grpc.Server{},
	}

	for i := 0; i < numStores; i++ {
		srv, addr, err := startStore()
		if err != nil {
			// Close so far started servers.
			st.Close()
			return nil, err
		}

		st.srvs[addr] = srv
	}

	return st, nil
}

func startStore() (*grpc.Server, string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", err
	}

	srv := grpc.NewServer()
	storepb.RegisterStoreServer(srv, &testStore{info: storepb.InfoResponse{
		Labels: []storepb.Label{
			{
				Name:  "addr",
				Value: listener.Addr().String(),
			},
		},
	}})
	go func() {
		srv.Serve(listener)
	}()

	return srv, listener.Addr().String(), nil
}

func (s *testStores) StoreAddresses() []string {
	var stores []string
	for addr := range s.srvs {
		stores = append(stores, addr)
	}
	return stores
}

func (s *testStores) Close() {
	for _, srv := range s.srvs {
		srv.Stop()
	}
	s.srvs = nil
}

func (s *testStores) CloseOne(addr string) {
	srv, ok := s.srvs[addr]
	if !ok {
		return
	}

	srv.Stop()
	delete(s.srvs, addr)
}

func specsFromAddrFunc(addrs []string) func() []StoreSpec {
	return func() (specs []StoreSpec) {
		for _, addr := range addrs {
			specs = append(specs, NewStaticStoreSpec(addr))
		}
		return specs
	}
}

func TestStoreSet_AllAvailable_ThenDown(t *testing.T) {
	defer leaktest.CheckTimeout(t, 10*time.Second)()

	st, err := newTestStores(2)
	testutil.Ok(t, err)
	defer st.Close()

	initialStoreAddr := st.StoreAddresses()

	// Testing if duplicates can cause weird results.
	initialStoreAddr = append(initialStoreAddr, initialStoreAddr[0])
	storeSet := NewStoreSet(nil, nil, specsFromAddrFunc(initialStoreAddr), testGRPCOpts)
	storeSet.gRPCRetryTimeout = 2 * time.Second
	defer storeSet.Close()

	// Should not matter how many of these we run.
	storeSet.Update(context.Background())
	storeSet.Update(context.Background())

	testutil.Assert(t, len(storeSet.stores) == 2, "all services should respond just fine, so we expect all clients to be ready.")

	for addr, store := range storeSet.stores {
		testutil.Equals(t, addr, store.addr)
		testutil.Equals(t, 1, len(store.labels))
		testutil.Equals(t, "addr", store.labels[0].Name)
		testutil.Equals(t, addr, store.labels[0].Value)
	}

	st.CloseOne(initialStoreAddr[0])

	// We expect Update to tear down store client for closed store server.
	storeSet.Update(context.Background())

	testutil.Assert(t, len(storeSet.stores) == 1, "only one service should respond just fine, so we expect one client to be ready.")

	addr := initialStoreAddr[1]
	store, ok := storeSet.stores[addr]
	testutil.Assert(t, ok, "addr exist")
	testutil.Equals(t, addr, store.addr)
	testutil.Equals(t, 1, len(store.labels))
	testutil.Equals(t, "addr", store.labels[0].Name)
	testutil.Equals(t, addr, store.labels[0].Value)

}

func TestStoreSet_StaticStores_OneAvailable(t *testing.T) {
	defer leaktest.CheckTimeout(t, 10*time.Second)()

	st, err := newTestStores(2)
	testutil.Ok(t, err)
	defer st.Close()

	initialStoreAddr := st.StoreAddresses()
	st.CloseOne(initialStoreAddr[0])

	storeSet := NewStoreSet(nil, nil, specsFromAddrFunc(initialStoreAddr), testGRPCOpts)
	storeSet.gRPCRetryTimeout = 2 * time.Second
	defer storeSet.Close()

	// Should not matter how many of these we run.
	storeSet.Update(context.Background())
	storeSet.Update(context.Background())

	testutil.Assert(t, len(storeSet.stores) == 1, "only one service should respond just fine, so we expect one client to be ready.")

	addr := initialStoreAddr[1]
	store, ok := storeSet.stores[addr]
	testutil.Assert(t, ok, "addr exist")
	testutil.Equals(t, addr, store.addr)
	testutil.Equals(t, 1, len(store.labels))
	testutil.Equals(t, "addr", store.labels[0].Name)
	testutil.Equals(t, addr, store.labels[0].Value)
}

func TestStoreSet_StaticStores_NoneAvailable(t *testing.T) {
	defer leaktest.CheckTimeout(t, 10*time.Second)()

	st, err := newTestStores(2)
	testutil.Ok(t, err)
	defer st.Close()

	initialStoreAddr := st.StoreAddresses()
	st.CloseOne(initialStoreAddr[0])
	st.CloseOne(initialStoreAddr[1])

	storeSet := NewStoreSet(nil, nil, specsFromAddrFunc(initialStoreAddr), testGRPCOpts)
	storeSet.gRPCRetryTimeout = 2 * time.Second

	// Should not matter how many of these we run.
	storeSet.Update(context.Background())
	storeSet.Update(context.Background())
	testutil.Assert(t, len(storeSet.stores) == 0, "none of services should respond just fine, so we expect no client to be ready.")

	// Leak test will ensure that we don't keep client connection around.
}
