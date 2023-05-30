{{/*
Expand the name of the chart.
*/}}
{{- define "zitadel.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "zitadel.fullname" -}}
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
{{- define "zitadel.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "zitadel.labels" -}}
helm.sh/chart: {{ include "zitadel.chart" . }}
{{ include "zitadel.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "zitadel.selectorLabels" -}}
app.kubernetes.io/name: {{ include "zitadel.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "zitadel.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "zitadel.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create copy command or empty string
*/}}
{{- define "zitadel.makecpcommand" -}}
{{- if .value -}}
{{ printf "cp %s /chowned-secrets/" .path }}
{{- end -}}
{{- end -}}

{{/*
Join copy commands
*/}}
{{- define "zitadel.joincpcommands" -}}
{{- $cmd := "" }}
    {{- range .commands -}}
        {{- if . -}}
            {{- $cmd = printf "%s && %s" ( default "yes" . ) $cmd  -}}
        {{- end -}}
    {{- end -}}
{{ print $cmd }}
{{- end -}}

{{/*
Returns true if the full path is defined and the value at the end of the path is not empty
*/}}
{{- define "deepCheck" -}}
  {{- if eq (len .path) 0 -}}
    {{- if and .root (not (empty .root)) -}}
      {{- true -}}
    {{- end -}}
  {{- else -}}
    {{- $head := index .path 0 -}}
    {{- $tail := slice .path 1 (len .path) -}}
    {{- if hasKey .root $head -}}
      {{- include "deepCheck" (dict "root" (index .root $head) "path" $tail) -}}
    {{- end -}}
  {{- end -}}
{{- end -}}