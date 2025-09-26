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

{{- define "login.nginxPort" -}}
{{/*
This helper defines the port that the nginx sidecar listens on for incoming
traffic from the Service. It automatically selects 80 for HTTP or 443 for
HTTPS based on whether login.serverSslCrtSecret is configured. This is purely
for reusability across templates. The nginx sidecar MUST always listen on
either 80 or 443 (not user-configurable). The actual NextJS login app always
runs on port 3000 internally (also not user-configurable) - nginx proxies to it.
*/}}
{{- if .Values.login.serverSslCrtSecret -}}
443
{{- else -}}
80
{{- end -}}
{{- end -}}

{{/*
This helper defines the port that the NextJS login application listens on
internally within the pod. This is hardcoded to 3000 and is NOT user-configurable.
The nginx sidecar proxies requests from login.nginxPort to this port. This helper
exists purely for consistency and to avoid magic numbers throughout the templates.
*/}}
{{- define "login.nextjsPort" -}}
3000
{{- end -}}

{{- define "login.appProtocol" -}}
{{/*
This helper defines and validates the Kubernetes appProtocol for the Service port.
It automatically returns "kubernetes.io/https" when login.serverSslCrtSecret is
configured, otherwise "kubernetes.io/http". If the user explicitly sets
.Values.login.service.appProtocol, it validates that it matches the TLS config
and fails deployment if there's a mismatch. This ensures the appProtocol always
accurately reflects whether TLS is enabled.
*/}}
{{- $appProtocol := .Values.login.service.appProtocol -}}
{{- if $appProtocol -}}
  {{- if and $appProtocol (ne $appProtocol "kubernetes.io/http") (ne $appProtocol "kubernetes.io/https") -}}
    {{- fail (printf "login.service.appProtocol must be either 'kubernetes.io/http' or 'kubernetes.io/https', got: %s" $appProtocol) -}}
  {{- end -}}
  {{- if .Values.login.serverSslCrtSecret -}}
    {{- if ne $appProtocol "kubernetes.io/https" -}}
      {{- fail "login.service.appProtocol must be 'kubernetes.io/https' when login.serverSslCrtSecret is provided" -}}
    {{- end -}}
  {{- else -}}
    {{- if eq $appProtocol "kubernetes.io/https" -}}
      {{- fail "login.service.appProtocol cannot be 'kubernetes.io/https' when login.serverSslCrtSecret is not provided" -}}
    {{- end -}}
  {{- end -}}
{{ $appProtocol }}
{{- else if .Values.login.serverSslCrtSecret -}}
kubernetes.io/https
{{- else -}}
kubernetes.io/http
{{- end -}}
{{- end -}}

{{- define "login.servicePort" -}}
{{/*
This helper defines the externally accessible port on the Service where clients
connect. If .Values.login.service.port is explicitly set, it uses that value
for legacy compatibility. Otherwise, it automatically defaults to 443 for HTTPS
or 80 for HTTP based on whether login.serverSslCrtSecret is configured. This
ensures sensible defaults while preserving backward compatibility for existing
configs. Validates that port selection matches TLS configuration to prevent
mismatches.
*/}}
{{- $port := .Values.login.service.port -}}
{{- if $port -}}
  {{- if and .Values.login.serverSslCrtSecret (eq ($port | int) 80) -}}
    {{- fail "login.service.port is set to 80 but login.serverSslCrtSecret is configured. Use port 443 for HTTPS or remove login.serverSslCrtSecret" -}}
  {{- end -}}
  {{- if and (not .Values.login.serverSslCrtSecret) (eq ($port | int) 443) -}}
    {{- fail "login.service.port is set to 443 but login.serverSslCrtSecret is not configured. Use port 80 for HTTP or configure login.serverSslCrtSecret" -}}
  {{- end -}}
{{ $port }}
{{- else if .Values.login.serverSslCrtSecret -}}
443
{{- else -}}
80
{{- end -}}
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
