{ "items": [
{{- if and .Values.rules.specifiedInValues .Values.rules.value }}
  {
    "key": "{{ .Release.Namespace }}/prometheus-{{ .Release.Name }}-rules",
    "checksum": "0000000000000000000000000000000000000000000000000000000000000000"
  }
{{- end }}
]}