{{/*
Standard secret template for weAlist services
Usage in service chart:
  {{- include "wealist-common.secret" . }}
*/}}
{{- define "wealist-common.secret" -}}
{{- if .Values.secrets }}
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "wealist-common.fullname" . }}-secret
  labels:
    {{- include "wealist-common.labels" . | nindent 4 }}
type: Opaque
data:
  {{- range $key, $value := .Values.secrets }}
  {{- if $value }}
  {{ $key }}: {{ $value | b64enc | quote }}
  {{- end }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
External secret reference
For use with External Secrets Operator
*/}}
{{- define "wealist-common.externalSecret" -}}
{{- if .Values.externalSecrets }}
{{- if .Values.externalSecrets.enabled }}
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: {{ include "wealist-common.fullname" . }}-external-secret
  labels:
    {{- include "wealist-common.labels" . | nindent 4 }}
spec:
  refreshInterval: {{ .Values.externalSecrets.refreshInterval | default "1h" }}
  secretStoreRef:
    name: {{ .Values.externalSecrets.secretStore }}
    kind: {{ .Values.externalSecrets.secretStoreKind | default "SecretStore" }}
  target:
    name: {{ include "wealist-common.fullname" . }}-secret
    creationPolicy: Owner
  data:
    {{- range .Values.externalSecrets.data }}
    - secretKey: {{ .secretKey }}
      remoteRef:
        key: {{ .remoteKey }}
        {{- if .property }}
        property: {{ .property }}
        {{- end }}
    {{- end }}
{{- end }}
{{- end }}
{{- end }}
