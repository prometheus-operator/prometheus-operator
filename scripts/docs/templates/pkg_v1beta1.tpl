{{ define "packages" }}
---
title: "Monitoring v1beta1 API Reference"
description: "Generated API reference for monitoring.coreos.com/v1beta1"
draft: false
images: []
menu: "operator"
weight: 154
toc: true
---

> This page is automatically generated with `gen-crd-api-reference-docs`.

{{ $pkg := index .packages 0 }}

<h2 id="{{- packageAnchorID $pkg -}}">
    {{- packageDisplayName $pkg -}}
</h2>

{{ with (index $pkg.GoPackages 0 )}}
    {{ with .DocComments }}
    <div>
        {{ safe (renderComments .) }}
    </div>
    {{ end }}
{{ end }}

Resource Types:
<ul>
{{- range (visibleTypes (sortedTypes $pkg.Types)) -}}
    {{ if isExportedType . -}}
    <li>
        <a href="{{ linkForType . }}">{{ typeDisplayName . }}</a>
    </li>
    {{- end }}
{{- end -}}
</ul>

{{ range (visibleTypes (sortedTypes $pkg.Types))}}
    {{ template "type" .  }}
{{ end }}
<hr/>
{{ end }}
