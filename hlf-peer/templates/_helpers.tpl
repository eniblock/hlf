{{/*
Expand the name of the chart.
*/}}
{{- define "hlf-peer.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "hlf-peer.fullname" -}}
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
{{- define "hlf-peer.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "hlf-peer.labels" -}}
helm.sh/chart: {{ include "hlf-peer.chart" . }}
{{ include "hlf-peer.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "hlf-peer.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hlf-peer.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "hlf-peer.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "hlf-peer.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create secret data with automatic initialization
Parameter: [$, Secret name, key, length, user value path]
*/}}
{{- define "hlf-peer.automaticSecret" -}}
{{- $ := index . 0 -}}
{{- $name := index . 1 -}}
{{- $key := index . 2 -}}
{{- $secretLength := int (index . 3) }}
{{- $userValue := index . 4 }}
{{- if $userValue }}
  {{ $key }}: {{ $userValue | b64enc | quote }}
{{- else if ($.Values.global).dev }}
  {{ $key }}: {{ printf "%s-%s" $name $key | sha256sum | trunc $secretLength | b64enc | quote }}
{{- else if and ($.Release.IsUpgrade) (lookup "v1" "Secret" $.Release.Namespace $name) }}
  {{ $key }}: {{ index (lookup "v1" "Secret" $.Release.Namespace $name).data $key }}
{{- else }}
  {{ $key }}: {{ randAlphaNum $secretLength | b64enc | quote }}
{{- end }}
{{- end }}
