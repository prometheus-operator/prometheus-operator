package main

import (
	"flag"
	"github.com/golang/glog"
	"k8s.io/klog"
)

type glogWriter struct{}

func (file *glogWriter) Write(data []byte) (n int, err error) {
	glog.InfoDepth(0, string(data))
	return len(data), nil
}

func main() {
	flag.Set("alsologtostderr", "true")

	var flags flag.FlagSet
	klog.InitFlags(&flags)
	flags.Set("skip_headers", "true")
	flag.Parse()

	klog.SetOutput(&glogWriter{})
	klog.Info("nice to meet you")
	glog.Flush()
	klog.Flush()
}
