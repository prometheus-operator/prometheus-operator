package v1

// PodDNSConfig defines the DNS parameters of a pod in addition to
// those generated from DNSPolicy.
type PodDNSConfig struct {
	// A list of DNS name server IP addresses.
	// This will be appended to the base nameservers generated from DNSPolicy.
	// Duplicated nameservers will be removed.
	// +kubebuilder:validation:Optional
	Nameservers []string `json:"nameservers,omitempty"`

	// A list of DNS search domains for host-name lookup.
	// This will be appended to the base search paths generated from DNSPolicy.
	// Duplicated search paths will be removed.
	// +kubebuilder:validation:Optional
	Searches []string `json:"searches,omitempty"`

	// A list of DNS resolver options.
	// This will be merged with the base options generated from DNSPolicy.
	// Duplicated entries will be removed. Resolution options given in Options
	// will override those that appear in the base DNSPolicy.
	// +kubebuilder:validation:Optional
	// +listType=map
	// +listMapKey=name
	Options []PodDNSConfigOption `json:"options,omitempty"`
}

// PodDNSConfigOption defines DNS resolver options of a pod.
type PodDNSConfigOption struct {
	// Name is required and must be unique.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Value is optional.
	// +kubebuilder:validation:Optional
	Value *string `json:"value,omitempty"`
}

// DNSPolicy specifies the DNS policy for the pod.
// +kubebuilder:validation:Enum=ClusterFirstWithHostNet;ClusterFirst;Default;None
type DNSPolicy string
