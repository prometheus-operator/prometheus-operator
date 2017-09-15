package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"encoding/json"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	AlertrulesKind = "Alertrule"
	AlertruleName  = "alertrules"
)

type AlertrulesGetter interface {
	Alertrules(namespace string) AlertruleInterface
}

type AlertruleInterface interface {
	Create(*Alertrule) (*Alertrule, error)
	Get(name string, opts metav1.GetOptions) (*Alertrule, error)
	Update(*Alertrule) (*Alertrule, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error
}

type alertrules struct {
	restClient rest.Interface
	client     *dynamic.ResourceClient
	ns         string
}

func (a *alertrules) Create(o *Alertrule) (*Alertrule, error) {
	ua, err := UnstructuredFromAlertrule(o)
	if err != nil {
		return nil, err
	}
	ua, err = a.client.Create(ua)
	if err != nil {
		return nil, err
	}
	return AlertruleFromUnstructured(ua)
}

func (a *alertrules) Update(o *Alertrule) (*Alertrule, error) {
	ua, err := UnstructuredFromAlertrule(o)
	if err != nil {
		return nil, err
	}

	cura, err := a.Get(o.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get current version for update")
	}
	ua.SetResourceVersion(cura.ObjectMeta.ResourceVersion)

	ua, err = a.client.Update(ua)
	if err != nil {
		return nil, err
	}

	return AlertruleFromUnstructured(ua)
}

func (a *alertrules) Get(name string, opts metav1.GetOptions) (*Alertrule, error) {
	obj, err := a.client.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return AlertruleFromUnstructured(obj)
}

func (a *alertrules) Delete(name string, options *metav1.DeleteOptions) error {
	return a.client.Delete(name, options)
}

func (a *alertrules) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := a.restClient.Get().
		Namespace(a.ns).
		Resource("alertrules").
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		return nil, err
	}
	var p AlertruleList
	return &p, json.Unmarshal(b, &p)
}

func (a *alertrules) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	r, err := a.restClient.Get().
		Prefix("watch").
		Namespace(a.ns).
		Resource("alertrules").
		FieldsSelectorParam(nil).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&alertruleDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil
}

func (a *alertrules) DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error {
	return a.client.DeleteCollection(dopts, lopts)
}

func newAlertrules(r rest.Interface, c *dynamic.Client, namespace string) *alertrules {
	return &alertrules{
		r,
		c.Resource(
			&metav1.APIResource{
				Kind:       AlertrulesKind,
				Name:       AlertruleName,
				Namespaced: true,
			},
			namespace,
		),
		namespace,
	}
}

func UnstructuredFromAlertrule(a *Alertrule) (*unstructured.Unstructured, error) {
	a.TypeMeta.Kind = AlertrulesKind
	a.TypeMeta.APIVersion = Group + "/" + Version
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	var r unstructured.Unstructured
	if err := json.Unmarshal(b, &r.Object); err != nil {
		return nil, err
	}
	return &r, nil
}

func AlertruleFromUnstructured(r *unstructured.Unstructured) (*Alertrule, error) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		return nil, err
	}
	var a Alertrule
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	a.TypeMeta.Kind = AlertrulesKind
	a.TypeMeta.APIVersion = Group + "/" + Version
	return &a, nil
}

type alertruleDecoder struct {
	dec *json.Decoder
	close func() error
}

func (d *alertruleDecoder) Close() {
	d.close()
}

func (d *alertruleDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type watch.EventType
		Object Alertrule
	}

	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}