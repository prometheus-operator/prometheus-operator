package framework

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"
)

type TestCtx struct {
	Id         string
	cleanUpFns []finalizerFn
}

type finalizerFn func() error

func NewTestCtx(t *testing.T) TestCtx {
	prefix := strings.TrimPrefix(strings.ToLower(t.Name()), "test")

	id := prefix + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	return TestCtx{
		Id: id,
	}
}

func (ctx *TestCtx) BasicSetup(t *testing.T, kubeClient kubernetes.Interface) {
	if _, err := CreateNamespace(kubeClient, ctx.Id); err != nil {
		t.Fatal(err)
	}

	namespaceFinalizerFn := func() error {
		if err := DeleteNamespace(kubeClient, ctx.Id); err != nil {
			return err
		}
		return nil
	}

	ctx.AddFinalizerFn(namespaceFinalizerFn)
}

func (ctx *TestCtx) CleanUp(t *testing.T) {
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
