package v1

import "sort"

type ParamEntry struct {
	// name is the parameter name.
	// +required
	Name string `json:"name"`
	// values is the parameter values.
	// +optional
	Values []string `json:"values,omitempty"`
}

func (e *ParamEntry) ToMapEntry() (string, []string) { return e.Name, e.Values }

func (p *PodMetricsEndpoint) ParamsMap() map[string][]string {
	m := make(map[string][]string, len(p.Params))
	for _, e := range p.Params {
		m[e.Name] = e.Values
	}
	return m
}

func (p *PodMetricsEndpoint) SetParamsMap(m map[string][]string) {
	if len(m) == 0 {
		p.Params = nil
		return
	}
	entries := make([]ParamEntry, 0, len(m))
	for k, v := range m {
		entries = append(entries, ParamEntry{Name: k, Values: v})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	p.Params = entries
}
