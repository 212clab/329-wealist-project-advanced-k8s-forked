{{/*
Standard configmap template for weAlist services
Usage in service chart:
  {{- include "wealist-common.configmap" . }}
*/}}
{{- define "wealist-common.configmap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "wealist-common.fullname" . }}-config
  labels:
    {{- include "wealist-common.labels" . | nindent 4 }}
data:
  {{- range $key, $value := .Values.config }}
  {{ $key }}: {{ $value | quote }}
  {{- end }}
{{- end }}
