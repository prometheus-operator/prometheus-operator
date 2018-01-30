// +build ignore

package main

import (
	"bytes"
	"go/format"
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type Resource struct {
	GoType string
	// Plural name of the resource. If empty, the GoType lowercased + "s".
	Name  string
	Flags uint8
}

const (
	// Is the resource cluster scoped (e.g. "nodes")?
	NotNamespaced uint8 = 1 << iota
	// Many "review" resources can be created but not listed
	NoList uint8 = 1 << iota
)

type APIGroup struct {
	Package  string
	Group    string
	Versions map[string][]Resource
}

func init() {
	for _, group := range apiGroups {
		for _, resources := range group.Versions {
			for i, r := range resources {
				if r.Name == "" {
					r.Name = strings.ToLower(r.GoType) + "s"
				}
				resources[i] = r
			}
		}
	}
}

var apiGroups = []APIGroup{
	{
		Package: "admissionregistration",
		Group:   "admissionregistration.k8s.io",
		Versions: map[string][]Resource{
			"v1beta1": []Resource{
				{"MutatingWebhookConfiguration", "", NotNamespaced},
				{"ValidatingWebhookConfiguration", "", NotNamespaced},
			},
			"v1alpha1": []Resource{
				{"InitializerConfiguration", "", NotNamespaced},
			},
		},
	},
	{
		Package: "apiextensions",
		Group:   "apiextensions.k8s.io",
		Versions: map[string][]Resource{
			"v1beta1": []Resource{
				{"CustomResourceDefinition", "", NotNamespaced},
			},
		},
	},
	{
		Package: "apps",
		Group:   "apps",
		Versions: map[string][]Resource{
			"v1": []Resource{
				{"ControllerRevision", "", 0},
				{"DaemonSet", "", 0},
				{"Deployment", "", 0},
				{"ReplicaSet", "", 0},
				{"StatefulSet", "", 0},
			},
			"v1beta2": []Resource{
				{"ControllerRevision", "", 0},
				{"DaemonSet", "", 0},
				{"Deployment", "", 0},
				{"ReplicaSet", "", 0},
				{"StatefulSet", "", 0},
			},
			"v1beta1": []Resource{
				{"ControllerRevision", "", 0},
				{"Deployment", "", 0},
				{"StatefulSet", "", 0},
			},
		},
	},
	{
		Package: "authentication",
		Group:   "authentication.k8s.io",
		Versions: map[string][]Resource{
			"v1": []Resource{
				{"TokenReview", "", NotNamespaced | NoList},
			},
			"v1beta1": []Resource{
				{"TokenReview", "", NotNamespaced | NoList},
			},
		},
	},
	{
		Package: "authorization",
		Group:   "authorization.k8s.io",
		Versions: map[string][]Resource{
			"v1": []Resource{
				{"LocalSubjectAccessReview", "", NoList},
				{"SelfSubjectAccessReview", "", NotNamespaced | NoList},
				{"SelfSubjectRulesReview", "", NotNamespaced | NoList},
				{"SubjectAccessReview", "", NotNamespaced | NoList},
			},
			"v1beta1": []Resource{
				{"LocalSubjectAccessReview", "", NoList},
				{"SelfSubjectAccessReview", "", NotNamespaced | NoList},
				{"SelfSubjectRulesReview", "", NotNamespaced | NoList},
				{"SubjectAccessReview", "", NotNamespaced | NoList},
			},
		},
	},
	{
		Package: "autoscaling",
		Group:   "autoscaling",
		Versions: map[string][]Resource{
			"v1": []Resource{
				{"HorizontalPodAutoscaler", "", 0},
			},
			"v2beta1": []Resource{
				{"HorizontalPodAutoscaler", "", 0},
			},
		},
	},
	{
		Package: "batch",
		Group:   "batch",
		Versions: map[string][]Resource{
			"v1": []Resource{
				{"Job", "", 0},
			},
			"v1beta1": []Resource{
				{"CronJob", "", 0},
			},
			"v2alpha1": []Resource{
				{"CronJob", "", 0},
			},
		},
	},
	{
		Package: "certificates",
		Group:   "certificates.k8s.io",
		Versions: map[string][]Resource{
			"v1beta1": []Resource{
				{"CertificateSigningRequest", "", NotNamespaced},
			},
		},
	},
	{
		Package: "core",
		Group:   "",
		Versions: map[string][]Resource{
			"v1": []Resource{
				{"ComponentStatus", "componentstatuses", NotNamespaced},
				{"ConfigMap", "", 0},
				{"Endpoints", "endpoints", 0},
				{"LimitRange", "", 0},
				{"Namespace", "", NotNamespaced},
				{"Node", "", NotNamespaced},
				{"PersistentVolumeClaim", "", 0},
				{"PersistentVolume", "", NotNamespaced},
				{"Pod", "", 0},
				{"ReplicationController", "", 0},
				{"ResourceQuota", "", 0},
				{"Secret", "", 0},
				{"Service", "", 0},
				{"ServiceAccount", "", 0},
			},
		},
	},
	{
		Package: "events",
		Group:   "events.k8s.io",
		Versions: map[string][]Resource{
			"v1beta1": []Resource{
				{"Event", "", 0},
			},
		},
	},
	{
		Package: "extensions",
		Group:   "extensions",
		Versions: map[string][]Resource{
			"v1beta1": []Resource{
				{"DaemonSet", "", 0},
				{"Deployment", "", 0},
				{"Ingress", "ingresses", 0},
				{"NetworkPolicy", "networkpolicies", 0},
				{"PodSecurityPolicy", "podsecuritypolicies", NotNamespaced},
				{"ReplicaSet", "", 0},
			},
		},
	},
	{
		Package: "networking",
		Group:   "networking.k8s.io",
		Versions: map[string][]Resource{
			"v1": []Resource{
				{"NetworkPolicy", "networkpolicies", 0},
			},
		},
	},
	{
		Package: "policy",
		Group:   "policy",
		Versions: map[string][]Resource{
			"v1beta1": []Resource{
				{"PodDisruptionBudget", "", 0},
			},
		},
	},
	{
		Package: "rbac",
		Group:   "rbac.authorization.k8s.io",
		Versions: map[string][]Resource{
			"v1": []Resource{
				{"ClusterRole", "", NotNamespaced},
				{"ClusterRoleBinding", "", NotNamespaced},
				{"Role", "", 0},
				{"RoleBinding", "", 0},
			},
			"v1beta1": []Resource{
				{"ClusterRole", "", NotNamespaced},
				{"ClusterRoleBinding", "", NotNamespaced},
				{"Role", "", 0},
				{"RoleBinding", "", 0},
			},
			"v1alpha1": []Resource{
				{"ClusterRole", "", NotNamespaced},
				{"ClusterRoleBinding", "", NotNamespaced},
				{"Role", "", 0},
				{"RoleBinding", "", 0},
			},
		},
	},
	{
		Package: "scheduling",
		Group:   "scheduling.k8s.io",
		Versions: map[string][]Resource{
			"v1alpha1": []Resource{
				{"PriorityClass", "", NotNamespaced},
			},
		},
	},
	{
		Package: "settings",
		Group:   "settings.k8s.io",
		Versions: map[string][]Resource{
			"v1alpha1": []Resource{
				{"PodPreset", "", 0},
			},
		},
	},
	{
		Package: "storage",
		Group:   "storage.k8s.io",
		Versions: map[string][]Resource{
			"v1": []Resource{
				{"StorageClass", "", NotNamespaced},
			},
			"v1beta1": []Resource{
				{"StorageClass", "", NotNamespaced},
			},
			"v1alpha1": []Resource{
				{"VolumeAttachment", "", NotNamespaced},
			},
		},
	},
}

type templateData struct {
	Package   string
	Resources []templateResource
}

type templateResource struct {
	Group      string
	Version    string
	Name       string
	Type       string
	Namespaced bool
	List       bool
}

var tmpl = template.Must(template.New("").Parse(`package {{ .Package }}

import "github.com/ericchiang/k8s"

func init() {
	{{- range $i, $r := .Resources -}}
	k8s.Register("{{ $r.Group }}", "{{ $r.Version }}", "{{ $r.Name }}", {{ $r.Namespaced }}, &{{ $r.Type }}{})
	{{ end -}}
	{{- range $i, $r := .Resources -}}{{ if $r.List }}
	k8s.RegisterList("{{ $r.Group }}", "{{ $r.Version }}", "{{ $r.Name }}", {{ $r.Namespaced }}, &{{ $r.Type }}List{}){{ end -}}
	{{ end -}}
}
`))

func main() {
	for _, group := range apiGroups {
		for version, resources := range group.Versions {
			fp := filepath.Join("apis", group.Package, version, "register.go")
			data := templateData{Package: version}
			for _, r := range resources {
				data.Resources = append(data.Resources, templateResource{
					Group:      group.Group,
					Version:    version,
					Name:       r.Name,
					Type:       r.GoType,
					Namespaced: r.Flags&NotNamespaced == 0,
					List:       r.Flags&NoList == 0,
				})
			}

			buff := new(bytes.Buffer)
			if err := tmpl.Execute(buff, &data); err != nil {
				log.Fatal(err)
			}
			out, err := format.Source(buff.Bytes())
			if err != nil {
				log.Fatal(err)
			}
			if err := ioutil.WriteFile(fp, out, 0644); err != nil {
				log.Fatal(err)
			}
		}
	}
}
