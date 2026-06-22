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
{{- $tag := default .Chart.AppVersion .Values.image.tag }}
app.kubernetes.io/version: {{ (splitList "@" $tag | first) | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Login Labels
*/}}
{{- define "login.labels" -}}
helm.sh/chart: {{ include "zitadel.chart" . }}
{{ include "login.commonSelectorLabels" . }}
{{- $tag := default .Chart.AppVersion .Values.image.tag }}
app.kubernetes.io/version: {{ (splitList "@" $tag | first) | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{ include "componentSelectorLabel" "login" }}
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
Zitadel service labels
*/}}
{{- define "zitadel.service.labels" -}}
{{ include "zitadel.labels" . }}
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
Login component selector labels
*/}}
{{- define "login.selectorLabels" -}}
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
Return the pod security context for Zitadel workloads.
Prefers zitadel.podSecurityContext; falls back to the chart-wide podSecurityContext.
*/}}
{{- define "zitadel.podSecurityContext" -}}
{{- if .Values.zitadel.podSecurityContext }}
{{- toYaml .Values.zitadel.podSecurityContext -}}
{{- else }}
{{- toYaml (default (dict) .Values.podSecurityContext) -}}
{{- end }}
{{- end }}

{{/*
Return the container security context for Zitadel workloads.
Prefers zitadel.securityContext; falls back to the chart-wide securityContext.
*/}}
{{- define "zitadel.securityContext" -}}
{{- if .Values.zitadel.securityContext }}
{{- toYaml .Values.zitadel.securityContext -}}
{{- else }}
{{- toYaml (default (dict) .Values.securityContext) -}}
{{- end }}
{{- end }}

{{/*
Return the pod security context for Login workloads.
Prefers login.podSecurityContext; falls back to the chart-wide podSecurityContext.
*/}}
{{- define "login.podSecurityContext" -}}
{{- if .Values.login.podSecurityContext }}
{{- toYaml .Values.login.podSecurityContext -}}
{{- else }}
{{- toYaml (default (dict) .Values.podSecurityContext) -}}
{{- end }}
{{- end }}

{{/*
Return the container security context for Login workloads.
Prefers login.securityContext; falls back to the chart-wide securityContext.
*/}}
{{- define "login.securityContext" -}}
{{- if .Values.login.securityContext }}
{{- toYaml .Values.login.securityContext -}}
{{- else }}
{{- toYaml (default (dict) .Values.securityContext) -}}
{{- end }}
{{- end }}


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
Login service key Secret name
*/}}
{{- define "zitadel.loginServiceKeySecretName" -}}
{{- if .Values.login.loginServiceKeySecretName -}}
{{ .Values.login.loginServiceKeySecretName }}
{{- else -}}
{{ include "zitadel.fullname" . }}-login-service-key
{{- end -}}
{{- end -}}

{{/*
Admin service key Secret name. When zitadel.adminServiceKey.existingSecretName
is set, the operator supplies the keypair externally and the chart references it
verbatim. Otherwise the chart generates and manages a Secret of this name.
*/}}
{{- define "zitadel.adminServiceKeySecretName" -}}
{{- if .Values.zitadel.adminServiceKey.existingSecretName -}}
{{ .Values.zitadel.adminServiceKey.existingSecretName }}
{{- else -}}
{{ include "zitadel.fullname" . }}-admin-service-key
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

{{/*
Returns the path for the ZITADEL login liveness probe. This endpoint
checks the basic health of the login user interface service, ensuring it is
running and responsive without verifying deeper dependencies.
*/}}
{{- define "login.livenessProbePath" -}}
/ui/v2/login/healthy
{{- end -}}

{{/*
Returns the path for the ZITADEL login readiness probe. This endpoint
performs a more thorough check to verify that the service is fully ready to
accept user traffic and can connect to its required backend dependencies.
*/}}
{{- define "login.readinessProbePath" -}}
/ui/v2/login/ready
{{- end -}}

{{/*
Returns the path for the ZITADEL login startup probe. It uses the same
thorough check as the readiness probe to ensure the application is fully
initialized before its other probes (liveness, readiness) take over.
*/}}
{{- define "login.startupProbePath" -}}
/ui/v2/login/ready
{{- end -}}

{{/*
Returns the path for the ZITADEL liveness probe. This endpoint provides a
basic health check, confirming that the main ZITADEL process is running and
able to respond to requests.
*/}}
{{- define "zitadel.livenessProbePath" -}}
/debug/healthz
{{- end -}}

{{/*
Returns the path for the ZITADEL readiness probe. This is a more detailed
check that verifies the service is not only running but has also
successfully connected to its database and is ready to serve traffic.
*/}}
{{- define "zitadel.readinessProbePath" -}}
/debug/ready
{{- end -}}

{{/*
Returns the path for the ZITADEL startup probe. It uses the same readiness
check to allow the container sufficient time to complete its lengthy
initialization, especially connecting to the database, before other probes begin.
*/}}
{{- define "zitadel.startupProbePath" -}}
/debug/ready
{{- end -}}


{{/*
Env vars for DB-talking containers: auto-generates a bundled PostgreSQL DSN
when the subchart is enabled and no explicit Database.Postgres.Host is set,
then appends any user-supplied .Values.env entries.
*/}}
{{- define "zitadel.dbEnv" -}}
{{- $host := "" -}}
{{- if (((.Values.zitadel).configmapConfig).Database) -}}
  {{- $host = dig "Postgres" "Host" "" .Values.zitadel.configmapConfig.Database -}}
{{- end -}}
{{- if and .Values.postgresql.enabled (not $host) }}
{{- $pgHost := printf "%s-postgresql" .Release.Name -}}
{{- if .Values.postgresql.fullnameOverride -}}
{{- $pgHost = .Values.postgresql.fullnameOverride -}}
{{- else if .Values.postgresql.nameOverride -}}
{{- $pgHost = printf "%s-%s" .Release.Name .Values.postgresql.nameOverride -}}
{{- end }}
- name: ZITADEL_DATABASE_POSTGRES_DSN
  value: "host={{ $pgHost }} port=5432 user=postgres password={{ .Values.postgresql.auth.postgresPassword }} dbname={{ .Values.postgresql.auth.database }} sslmode=disable"
{{- end }}
{{- with .Values.env }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{/*
Build the effective configmap config. Injects SystemAPIUsers entries that
reference X.509 public certificates mounted into the ZITADEL container:
  - login-client (IAM_LOGIN_CLIENT) when login.enabled=true
  - admin-client (IAM_OWNER) when zitadel.adminServiceKey.enabled=true
The admin-client replaces the legacy imperatively-created IAM machine user:
the operator authenticates against the System API with the private key, so no
Secret is ever created at runtime. User-supplied configmapConfig values always
win, so an operator may override either entry (or add their own).
*/}}
{{- define "zitadel.mergedConfigmapConfig" -}}
{{- $config := deepCopy .Values.zitadel.configmapConfig -}}
{{- $sysUsers := dict -}}
{{- if .Values.login.enabled -}}
{{- $_ := set $sysUsers "login-client" (dict "Path" "/secrets/login-client/tls.crt" "Memberships" (list (dict "MemberType" "System" "Roles" (list "IAM_LOGIN_CLIENT")))) -}}
{{- end -}}
{{- if .Values.zitadel.adminServiceKey.enabled -}}
{{- $_ := set $sysUsers "admin-client" (dict "Path" "/secrets/admin-client/tls.crt" "Memberships" (list (dict "MemberType" "System" "Roles" (list "IAM_OWNER")))) -}}
{{- end -}}
{{- if $sysUsers -}}
{{- $config = mergeOverwrite (dict "SystemAPIUsers" $sysUsers) $config -}}
{{- end -}}
{{- $config | toYaml -}}
{{- end -}}

{{/*
Return the effective ingress className for the ZITADEL API ingress.
*/}}
{{- define "zitadel.ingressClassName" -}}
{{- if .Values.ingress.className -}}
{{- .Values.ingress.className -}}
{{- end -}}
{{- end -}}

{{/*
Return the effective ingress className for the ZITADEL Login ingress.
*/}}
{{- define "zitadel.login.ingressClassName" -}}
{{- if .Values.login.ingress.className -}}
{{- .Values.login.ingress.className -}}
{{- end -}}
{{- end -}}
