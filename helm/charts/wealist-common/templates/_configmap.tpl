{{/*
Standard configmap template for weAlist services
Usage in service chart:
  {{- include "wealist-common.configmap" . }}

Config merging priority (higher number = higher priority):
1. shared.config (from environment files - common for all services)
2. config (from service values.yaml - service-specific, overrides shared)

Note: Helm merge gives precedence to first arg, so we use mustMergeOverwrite
to ensure service-specific config overrides shared config.
*/}}
{{- define "wealist-common.configmap" -}}
{{- $mergedConfig := dict }}
{{- /* Start with shared config from environment files */}}
{{- if .Values.shared }}
{{- if .Values.shared.config }}
{{- $mergedConfig = .Values.shared.config }}
{{- end }}
{{- end }}
{{- /* Override with service-specific config */}}
{{- if .Values.config }}
{{- $mergedConfig = mustMergeOverwrite $mergedConfig .Values.config }}
{{- end }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "wealist-common.fullname" . }}-config
  labels:
    {{- include "wealist-common.labels" . | nindent 4 }}
data:
  {{- range $key, $value := $mergedConfig }}
  {{ $key }}: {{ $value | quote }}
  {{- end }}
{{- end }}
