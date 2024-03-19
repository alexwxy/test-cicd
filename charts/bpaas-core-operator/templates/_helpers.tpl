{{/*
Expand the name of the chart.
*/}}
{{- define "bpaas-core-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bpaas-core-operator.fullname" -}}
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
{{- define "bpaas-core-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bpaas-core-operator.labels" -}}
helm.sh/chart: {{ include "bpaas-core-operator.chart" . }}
{{ include "bpaas-core-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "bpaas-core-operator-apiserver.labels" -}}
helm.sh/chart: {{ include "bpaas-core-operator.chart" . }}
{{ include "bpaas-core-operator-apiserver.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bpaas-core-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bpaas-core-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "bpaas-core-operator-apiserver.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bpaas-core-operator.name" . }}-apiserver
app.kubernetes.io/instance: {{ .Release.Name }}-apiserver
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bpaas-core-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bpaas-core-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the default bpaas-core-operator app version
*/}}
{{- define "bpaas-core-operator.defaultTag" -}}
  {{- default .Chart.AppVersion .Values.images.tag }}
{{- end -}}
