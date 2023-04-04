{{/*
Expand the name of the chart.
*/}}
{{- define "money-tracker-telegram.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "money-tracker-telegram.fullname" -}}
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

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "money-tracker-telegram.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "money-tracker-telegram.labels" -}}
helm.sh/chart: {{ include "money-tracker-telegram.chart" . }}
{{ include "money-tracker-telegram.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "money-tracker-telegram.selectorLabels" -}}
app.kubernetes.io/name: {{ include "money-tracker-telegram.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Set extra configurations for the database to be autowired
*/}}
{{- define "money-tracker-telegram.dbConfig" -}}
dbConfig:
  host: {{ .Release.Name }}-postgresql
  user: {{ .Values.postgresql.postgresqlUsername }}
  password: {{ .Values.postgresql.postgresqlPassword }}
  dbName: {{ .Values.postgresql.postgresqlDatabase }}
  port: 5432
{{- end }}
