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

	ctx.cleanUpFns = append(ctx.cleanUpFns, func() error {
		if err := DeleteNamespace(kubeClient, ctx.Id); err != nil {
			return err
		}
		return nil
	})
}

func (ctx *TestCtx) CleanUp(t *testing.T) {
	var eg errgroup.Group

	// TODO: Should be a stack not a list
	for _, f := range ctx.cleanUpFns {
		eg.Go(f)
	}

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}
