{{/*
Expand the name of the chart.
*/}}
{{- define "wealist-common.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "wealist-common.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- printf "%s" $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "wealist-common.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "wealist-common.labels" -}}
helm.sh/chart: {{ include "wealist-common.chart" . }}
{{ include "wealist-common.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: wealist
{{- if .Values.global }}
{{- if .Values.global.environment }}
app.kubernetes.io/env: {{ .Values.global.environment }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "wealist-common.selectorLabels" -}}
app.kubernetes.io/name: {{ include "wealist-common.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Database URL constructor
Usage: {{ include "wealist-common.databaseURL" (dict "user" "user_service" "password" "pass" "host" "postgres" "port" "5432" "db" "wealist_user_db") }}
*/}}
{{- define "wealist-common.databaseURL" -}}
postgresql://{{ .user }}:{{ .password }}@{{ .host }}:{{ .port }}/{{ .db }}?sslmode=disable
{{- end }}

{{/*
Image name constructor
Combines global registry with image repository and tag
*/}}
{{- define "wealist-common.image" -}}
{{- if .Values.global }}
{{- if .Values.global.imageRegistry }}
{{- printf "%s/%s:%s" .Values.global.imageRegistry .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) }}
{{- else }}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) }}
{{- end }}
{{- else }}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) }}
{{- end }}
{{- end }}
