{{/*
Expand the name of the chart.
*/}}
{{- define "zitadel.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Login Name label suffix
*/}}
{{- define "zitadel.login.name" -}}
{{ include "zitadel.name" . | trunc 57 }}-login
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
Create a default fully qualified login app name.
We suffix zitadel.fullname with -login.
*/}}
{{- define "zitadel.login.fullname" -}}
{{ include "zitadel.fullname" . | trunc 57 | trimSuffix "-" }}-login
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
app.kubernetes.io/version: {{ (.Values.image.tag | default .Chart.AppVersion | split "@")._0 | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Login Labels
*/}}
{{- define "login.labels" -}}
helm.sh/chart: {{ include "zitadel.chart" . }}
{{ include "login.selectorLabels" . }}
app.kubernetes.io/version: {{ (.Values.image.tag | default .Chart.AppVersion | split "@")._0 | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Init component labels
*/}}
{{- define "zitadel.init.labels" -}}
{{ include "zitadel.labels" . }}
{{ include "componentSelectorLabels" "init" }}
{{- end }}

{{/*
Setup component labels
*/}}
{{- define "zitadel.setup.labels" -}}
{{ include "zitadel.labels" . }}
{{ include "componentSelectorLabels" "setup" }}
{{- end }}

{{/*
Start component labels
*/}}
{{- define "zitadel.start.labels" -}}
{{ include "zitadel.labels" . }}
{{ include "componentSelectorLabels" "start" }}
{{- end }}

{{/*
Debug component labels
*/}}
{{- define "zitadel.debug.labels" -}}
{{ include "zitadel.labels" . }}
{{ include "componentSelectorLabels" "debug" }}
{{- end }}

{{/*
Login component labels
*/}}
{{- define "zitadel.login.labels" -}}
{{ include "zitadel.labels" . }}
{{ include "componentSelectorLabels" "login" }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "zitadel.selectorLabels" -}}
app.kubernetes.io/name: {{ include "zitadel.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Login Selector labels
*/}}
{{- define "login.selectorLabels" -}}
app.kubernetes.io/name: {{ include "zitadel.login.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Component selector label
*/}}
{{- define "componentSelectorLabels" -}}
app.kubernetes.io/component: {{ . }}
{{- end }}

{{/*
Init component selector labels
*/}}
{{- define "zitadel.init.selectorLabels" -}}
{{ include "zitadel.selectorLabels" . }}
{{ include "componentSelectorLabels" "init" }}
{{- end }}

{{/*
Setup component selector labels
*/}}
{{- define "zitadel.setup.selectorLabels" -}}
{{ include "zitadel.selectorLabels" . }}
{{ include "componentSelectorLabels" "setup" }}
{{- end }}

{{/*
Start component selector labels
*/}}
{{- define "zitadel.start.selectorLabels" -}}
{{ include "zitadel.selectorLabels" . }}
{{ include "componentSelectorLabels" "start" }}
{{- end }}

{{/*
Debug component selector labels
*/}}
{{- define "zitadel.debug.selectorLabels" -}}
{{ include "zitadel.selectorLabels" . }}
{{ include "componentSelectorLabels" "debug" }}
{{- end }}

{{/*
Login component selector labels
*/}}
{{- define "zitade.login.selectorLabels" -}}
{{ include "login.selectorLabels" . }}
{{ include "componentSelectorLabels" "login" }}
{{- end }}


{{/*
Create the name of the zitadel service account to use
*/}}
{{- define "zitadel.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "zitadel.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the login service account to use
*/}}
{{- define "login.serviceAccountName" -}}
{{- if .Values.login.serviceAccount.create }}
{{- default (include "zitadel.login.fullname" .) .Values.login.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.login.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Returns true if the full path is defined and the value at the end of the path is not empty
*/}}
{{- define "deepCheck" -}}
  {{- if empty .root -}}
    {{/* Break early */}}
  {{- else if eq (len .path) 0 -}}
    {{- not (empty .root) -}}
  {{- else -}}
    {{- $head := index .path 0 -}}
    {{- $tail := slice .path 1 (len .path) -}}
    {{- if hasKey .root $head -}}
        {{- include "deepCheck" (dict "root" (index .root $head) "path" $tail) -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Returns the database config from the secretConfig or else from the configmapConfig
*/}}
{{- define "zitadel.dbconfig.json" -}}
    {{- if include "deepCheck" (dict "root" . "path" (splitList "." "Values.zitadel.secretConfig.Database")) -}}
    {{- .Values.zitadel.secretConfig.Database | toJson -}}
    {{- else if include "deepCheck" (dict "root" . "path" (splitList "." "Values.zitadel.configmapConfig.Database")) -}}
    {{- .Values.zitadel.configmapConfig.Database | toJson -}}
    {{- else -}}
    {{- dict | toJson -}}
    {{- end -}}
{{- end -}}

{{/*
Returns a dict with the databases key in the yaml and the environment variable part, either COCKROACH or POSTGRES, in uppercase letters.
*/}}
{{- define "zitadel.dbkey.json" -}}
  {{- range $i, $key := (include "zitadel.dbconfig.json" . | fromJson | keys ) -}}
    {{- if or (eq (lower $key) "postgres" ) (eq (lower $key) "pg" ) -}}
        {"key": "{{ $key }}", "env": "POSTGRES" }
    {{- else if or (eq (lower $key) "cockroach" ) (eq (lower $key) "crdb" ) -}}
        {"key": "{{ $key }}", "env": "COCKROACH" }
    {{- end -}}
  {{- end -}}
{{- end -}}
