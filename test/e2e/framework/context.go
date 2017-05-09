package framework

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

type TestCtx struct {
	ID         string
	cleanUpFns []finalizerFn
}

type finalizerFn func() error

func (f *Framework) NewTestCtx(t *testing.T) TestCtx {
	prefix := strings.TrimPrefix(strings.ToLower(t.Name()), "test")

	id := prefix + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	return TestCtx{
		ID: id,
	}
}

// GetObjID returns an ascending ID based on the length of cleanUpFns. It is
// based on the premise that every new object also appends a new finalizerFn on
// cleanUpFns. This can e.g. be used to create multiple namespaces in the same
// test context.
func (ctx *TestCtx) GetObjID() string {
	return ctx.ID + "-" + strconv.Itoa(len(ctx.cleanUpFns))
}

func (ctx *TestCtx) Cleanup(t *testing.T) {
	var eg errgroup.Group

	for i := len(ctx.cleanUpFns) - 1; i >= 0; i-- {
		eg.Go(ctx.cleanUpFns[i])
	}

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}

func (ctx *TestCtx) AddFinalizerFn(fn finalizerFn) {
	ctx.cleanUpFns = append(ctx.cleanUpFns, fn)
}
