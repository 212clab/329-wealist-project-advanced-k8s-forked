{{/*
Standard deployment template for weAlist services
Usage in service chart:
  {{- include "wealist-common.deployment" . }}
*/}}
{{- define "wealist-common.deployment" -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "wealist-common.fullname" . }}
  labels:
    {{- include "wealist-common.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount | default 1 }}
  selector:
    matchLabels:
      {{- include "wealist-common.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "wealist-common.selectorLabels" . | nindent 8 }}
      {{- if .Values.podAnnotations }}
      annotations:
        {{- toYaml .Values.podAnnotations | nindent 8 }}
      {{- end }}
    spec:
      {{- if .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml .Values.imagePullSecrets | nindent 8 }}
      {{- end }}
      containers:
        - name: {{ .Chart.Name }}
          image: {{ include "wealist-common.image" . | quote }}
          imagePullPolicy: {{ .Values.image.pullPolicy | default "Always" }}
          ports:
            - name: http
              containerPort: {{ .Values.service.targetPort }}
              protocol: TCP
          {{- if .Values.healthCheck }}
          {{- if .Values.healthCheck.liveness }}
          {{- if .Values.healthCheck.liveness.enabled }}
          livenessProbe:
            httpGet:
              path: {{ .Values.healthCheck.liveness.path }}
              port: {{ .Values.healthCheck.liveness.port | default .Values.service.targetPort }}
            initialDelaySeconds: {{ .Values.healthCheck.liveness.initialDelaySeconds | default 10 }}
            periodSeconds: {{ .Values.healthCheck.liveness.periodSeconds | default 10 }}
            timeoutSeconds: {{ .Values.healthCheck.liveness.timeoutSeconds | default 1 }}
            successThreshold: {{ .Values.healthCheck.liveness.successThreshold | default 1 }}
            failureThreshold: {{ .Values.healthCheck.liveness.failureThreshold | default 3 }}
          {{- end }}
          {{- end }}
          {{- if .Values.healthCheck.readiness }}
          {{- if .Values.healthCheck.readiness.enabled }}
          readinessProbe:
            httpGet:
              path: {{ .Values.healthCheck.readiness.path }}
              port: {{ .Values.healthCheck.readiness.port | default .Values.service.targetPort }}
            initialDelaySeconds: {{ .Values.healthCheck.readiness.initialDelaySeconds | default 5 }}
            periodSeconds: {{ .Values.healthCheck.readiness.periodSeconds | default 5 }}
            timeoutSeconds: {{ .Values.healthCheck.readiness.timeoutSeconds | default 1 }}
            successThreshold: {{ .Values.healthCheck.readiness.successThreshold | default 1 }}
            failureThreshold: {{ .Values.healthCheck.readiness.failureThreshold | default 3 }}
          {{- end }}
          {{- end }}
          {{- end }}
          {{- if .Values.resources }}
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
          {{- end }}
          {{- if .Values.envFrom }}
          envFrom:
            {{- toYaml .Values.envFrom | nindent 12 }}
          {{- end }}
          {{- if .Values.env }}
          env:
            {{- toYaml .Values.env | nindent 12 }}
          {{- end }}
          {{- if .Values.volumeMounts }}
          volumeMounts:
            {{- toYaml .Values.volumeMounts | nindent 12 }}
          {{- end }}
      {{- if .Values.volumes }}
      volumes:
        {{- toYaml .Values.volumes | nindent 8 }}
      {{- end }}
      {{- if .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml .Values.nodeSelector | nindent 8 }}
      {{- end }}
      {{- if .Values.affinity }}
      affinity:
        {{- toYaml .Values.affinity | nindent 8 }}
      {{- end }}
      {{- if .Values.tolerations }}
      tolerations:
        {{- toYaml .Values.tolerations | nindent 8 }}
      {{- end }}
{{- end }}
