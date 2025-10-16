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
{{ include "zitadel.commonSelectorLabels" . }}
app.kubernetes.io/version: {{ (.Values.image.tag | default .Chart.AppVersion | split "@")._0 | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Login Labels
*/}}
{{- define "login.labels" -}}
helm.sh/chart: {{ include "zitadel.chart" . }}
{{ include "login.commonSelectorLabels" . }}
app.kubernetes.io/version: {{ (.Values.image.tag | default .Chart.AppVersion | split "@")._0 | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Init component labels
*/}}
{{- define "zitadel.init.labels" -}}
{{ include "zitadel.labels" . }}
{{ include "componentSelectorLabel" "init" }}
{{- end }}

{{/*
Setup component labels
*/}}
{{- define "zitadel.setup.labels" -}}
{{ include "zitadel.labels" . }}
{{ include "componentSelectorLabel" "setup" }}
{{- end }}

{{/*
Start component labels
*/}}
{{- define "zitadel.start.labels" -}}
{{ include "zitadel.labels" . }}
{{ include "componentSelectorLabel" "start" }}
{{- end }}

{{/*
Debug component labels
*/}}
{{- define "zitadel.debug.labels" -}}
{{ include "zitadel.labels" . }}
{{ include "componentSelectorLabel" "debug" }}
{{- end }}

{{/*
Zitadel service labels
*/}}
{{- define "zitadel.service.labels" -}}
{{ include "zitadel.labels" . }}
{{- end }}

{{/*
Login component labels
*/}}
{{- define "zitadel.login.labels" -}}
{{ include "login.labels" . }}
{{ include "componentSelectorLabel" "login" }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "zitadel.commonSelectorLabels" -}}
app.kubernetes.io/name: {{ include "zitadel.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Login Selector labels
*/}}
{{- define "login.commonSelectorLabels" -}}
app.kubernetes.io/name: {{ include "zitadel.login.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Component selector label
*/}}
{{- define "componentSelectorLabel" -}}
app.kubernetes.io/component: {{ . }}
{{- end }}

{{/*
Init component selector labels
*/}}
{{- define "zitadel.init.selectorLabels" -}}
{{ include "zitadel.commonSelectorLabels" . }}
{{ include "componentSelectorLabel" "init" }}
{{- end }}

{{/*
Setup component selector labels
*/}}
{{- define "zitadel.setup.selectorLabels" -}}
{{ include "zitadel.commonSelectorLabels" . }}
{{ include "componentSelectorLabel" "setup" }}
{{- end }}

{{/*
Start component selector labels
*/}}
{{- define "zitadel.start.selectorLabels" -}}
{{ include "zitadel.commonSelectorLabels" . }}
{{ include "componentSelectorLabel" "start" }}
{{- end }}

{{/*
Debug component selector labels
*/}}
{{- define "zitadel.debug.selectorLabels" -}}
{{ include "zitadel.commonSelectorLabels" . }}
{{ include "componentSelectorLabel" "debug" }}
{{- end }}

{{/*
Login component selector labels
*/}}
{{- define "zitadel.login.selectorLabels" -}}
{{ include "login.commonSelectorLabels" . }}
{{ include "componentSelectorLabel" "login" }}
{{- end }}

{{/*
Zitadel Service Selector labels
*/}}
{{- define "zitadel.service.selectorLabels" -}}
{{ include "zitadel.commonSelectorLabels" . }}
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

{{- define "zitadel.containerPort" -}}
8080
{{- end -}}

{{- define "login.containerPort" -}}
3000
{{- end -}}

{{/*
ZITADEL config ConfigMap name
*/}}
{{- define "zitadel.configmapName" -}}
{{ include "zitadel.fullname" . }}-config-yaml
{{- end -}}

{{/*
Login config ConfigMap name
*/}}
{{- define "login.configmapName" -}}
{{ include "zitadel.login.fullname" . }}-config-dotenv
{{- end -}}

{{/*
ZITADEL secrets Secret name
*/}}
{{- define "zitadel.secretName" -}}
{{ include "zitadel.fullname" . }}-secrets-yaml
{{- end -}}

{{/*
ZITADEL masterkey Secret name
*/}}
{{- define "zitadel.masterkeySecretName" -}}
{{- if .Values.zitadel.masterkeySecretName -}}
{{ .Values.zitadel.masterkeySecretName }}
{{- else -}}
{{ include "zitadel.fullname" . }}-masterkey
{{- end -}}
{{- end -}}

{{/*
Database SSL CA certificate Secret name
*/}}
{{- define "zitadel.dbSslCaCrtSecretName" -}}
{{- if .Values.zitadel.dbSslCaCrtSecret -}}
{{ .Values.zitadel.dbSslCaCrtSecret }}
{{- else -}}
{{ include "zitadel.fullname" . }}-db-ssl-ca-crt
{{- end -}}
{{- end -}}

{{/*
Returns the internal cluster endpoint URL for ZITADEL health checks.
This is used by wait4x and other internal pod-to-pod communication.
The URL scheme (http/https) is determined by the TLS configuration:
- If zitadel.configmapConfig.TLS.Enabled is true, uses https://
- Otherwise, uses http://
The URL format is: <scheme>://<service-name>:<port>/debug/ready
Example outputs:
  - http://my-release-zitadel:8080/debug/ready
  - https://my-release-zitadel:8080/debug/ready
*/}}
{{- define "zitadel.clusterEndpoint" -}}
{{- if include "deepCheck" (dict "root" .Values "path" (splitList "." "zitadel.configmapConfig.TLS.Enabled")) -}}
https://{{ include "zitadel.fullname" . }}:{{ .Values.service.port }}/debug/ready
{{- else -}}
http://{{ include "zitadel.fullname" . }}:{{ .Values.service.port }}/debug/ready
{{- end -}}
{{- end -}}


{{/*
Returns the PostgreSQL TCP endpoint for wait4x health checks.
Extracts the database host and port from ZITADEL configuration.
Supports both "Postgres" and "postgres" key variations.
Format: tcp://<host>:<port>
Example: tcp://db-postgresql:5432
*/}}
{{- define "zitadel.postgresEndpoint" -}}
{{- with .Values.zitadel.configmapConfig.Database.Postgres -}}
  {{- if .Host }}
    {{- .Host }}:{{ .Port | default 5432 }}
  {{- end -}}
{{- end -}}
{{- end -}}

{{/*
This helper template takes the Kubernetes cluster's version string, which
can be complex (e.g., "v1.28.5+k3s1"), and returns a sanitized, clean
version string in the "MAJOR.MINOR.PATCH" format. This is crucial for
creating valid container image tags that won't fail on Kubernetes
distributions with non-standard versioning schemes.

Its logic first uses the `semver` function to parse the full version
string, intelligently separating the core version numbers from extra
suffixes. The `printf` function then rebuilds the string using only the
major, minor, and patch components, guaranteeing a clean and valid output.
*/}}
{{- define "zitadel.kubeVersion" -}}
{{- $version := semver .Capabilities.KubeVersion.Version -}}
{{- printf "%d.%d.%d" $version.Major $version.Minor $version.Patch -}}
{{- end -}}
