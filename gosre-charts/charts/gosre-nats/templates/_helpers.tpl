{{- define "gosre-nats.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "gosre-nats.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "gosre-nats.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "gosre-nats.labels" -}}
helm.sh/chart: {{ include "gosre-nats.chart" . }}
{{ include "gosre-nats.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "gosre-nats.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gosre-nats.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/* Render nats.conf content — called with the root context */}}
{{- define "gosre-nats.config" }}
http_port: {{ .Values.service.monitorPort }}

jetstream {
  store_dir: /data/jetstream
  max_memory_store: {{ .Values.jetstream.maxMemoryStore }}
  max_file_store: {{ .Values.jetstream.maxFileStore }}
}
{{- if gt (int .Values.replicaCount) 1 }}

cluster {
  name: {{ .Release.Name }}-nats
  port: {{ .Values.service.clusterPort }}
  routes = [
    {{- range until (int .Values.replicaCount) }}
    nats-route://{{ include "gosre-nats.fullname" $ }}-{{ . }}.{{ include "gosre-nats.fullname" $ }}-headless:{{ $.Values.service.clusterPort }}
    {{- end }}
  ]
}
{{- end }}
{{- end }}
