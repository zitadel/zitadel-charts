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
Cleanup component labels
*/}}
{{- define "zitadel.cleanup.labels" -}}
{{ include "zitadel.labels" . }}
{{ include "componentSelectorLabel" "cleanup" }}
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

{{/*
Returns the database config from the secretConfig or else from the configmapConfig
*/}}
{{- define "zitadel.dbconfig.json" -}}
    {{- if (((.Values.zitadel).secretConfig).Database) -}}
    {{- .Values.zitadel.secretConfig.Database | toJson -}}
    {{- else if (((.Values.zitadel).configmapConfig).Database) -}}
    {{- .Values.zitadel.configmapConfig.Database | toJson -}}
    {{- else -}}
    {{- dict | toJson -}}
    {{- end -}}
{{- end -}}

{{/*
Returns a dict with the databases key in the yaml and the environment variable part, either COCKROACH or POSTGRES, in uppercase letters.
Defaults to POSTGRES when no Database section is present in the chart values (i.e. DSN mode).
*/}}
{{- define "zitadel.dbkey.json" -}}
  {{- $found := false -}}
  {{- range $i, $key := (include "zitadel.dbconfig.json" . | fromJson | keys ) -}}
    {{- if or (eq (lower $key) "postgres" ) (eq (lower $key) "pg" ) -}}
        {"key": "{{ $key }}", "env": "POSTGRES" }
        {{- $found = true -}}
    {{- else if or (eq (lower $key) "cockroach" ) (eq (lower $key) "crdb" ) -}}
        {"key": "{{ $key }}", "env": "COCKROACH" }
        {{- $found = true -}}
    {{- end -}}
  {{- end -}}
  {{- if not $found -}}
    {"key": "Postgres", "env": "POSTGRES" }
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
{{- if ((((.Values.zitadel).configmapConfig).TLS).Enabled) -}}
https://{{ include "zitadel.fullname" . }}:{{ .Values.service.port }}/debug/ready
{{- else -}}
http://{{ include "zitadel.fullname" . }}:{{ .Values.service.port }}/debug/ready
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
/ui/v2/login/security
{{- end -}}

{{/*
Returns the path for the ZITADEL login startup probe. It uses the same
thorough check as the readiness probe to ensure the application is fully
initialized before its other probes (liveness, readiness) take over.
*/}}
{{- define "login.startupProbePath" -}}
/ui/v2/login/security
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
Return the image for the machinekeyWriter (Standardized kubectl image).
Backward Compatibility Logic:
1. IF the legacy "setupJob.machinekeyWriter.image.repository" is set, use it (Legacy Mode).
2. ELSE use the new "tools.kubectl.image" with Global Registry support (New Mode).
*/}}
{{- define "kubectl.image" -}}
{{- /* Safely check if the legacy value exists without crashing on nil pointers */ -}}
{{- $legacyRepo := "" -}}
{{- if .Values.setupJob -}}
  {{- if .Values.setupJob.machinekeyWriter -}}
    {{- if .Values.setupJob.machinekeyWriter.image -}}
      {{- $legacyRepo = .Values.setupJob.machinekeyWriter.image.repository -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- if $legacyRepo -}}
  {{- /* 1. Legacy Mode: Use specific config, ignore global registry */ -}}
  {{- $tag := .Values.setupJob.machinekeyWriter.image.tag | default (include "zitadel.kubeVersion" .) -}}
  {{- printf "%s:%s" $legacyRepo $tag -}}
{{- else -}}
  {{- /* 2. New Mode: Use tools.kubectl with Global Registry */ -}}
  {{- /* Uses fully qualified image names for CRI-O v1.34+ compatibility */ -}}
  {{- $registry := .Values.imageRegistry | default "docker.io" -}}
  {{- $repo := .Values.tools.kubectl.image.repository | default "alpine/k8s" -}}
  {{- $tag := .Values.tools.kubectl.image.tag | default (include "zitadel.kubeVersion" .) -}}
  {{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- end -}}
{{- end -}}

{{/*
Return the image for the wait4x tool.
Uses fully qualified image names for CRI-O v1.34+ compatibility.
*/}}
{{- define "wait4x.image" -}}
{{- $registry := .Values.imageRegistry | default "docker.io" -}}
{{- $repo := .Values.tools.wait4x.image.repository | default "wait4x/wait4x" -}}
{{- $tag := .Values.tools.wait4x.image.tag | default "3.6" -}}
{{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- end -}}

{{/*
Return the PostgreSQL service hostname for the bundled subchart.
Respects fullnameOverride and nameOverride on the postgresql dependency.
*/}}
{{- define "zitadel.postgresqlHost" -}}
{{- if .Values.postgresql.fullnameOverride -}}
{{- .Values.postgresql.fullnameOverride -}}
{{- else if .Values.postgresql.nameOverride -}}
{{- printf "%s-%s" .Release.Name .Values.postgresql.nameOverride -}}
{{- else -}}
{{- printf "%s-postgresql" .Release.Name -}}
{{- end -}}
{{- end -}}

{{/*
Auto-generate ZITADEL database configuration from the bundled PostgreSQL subchart.
Only used in legacy "configmap mode" when the user has explicitly set
zitadel.configmapConfig.Database.Postgres.Host. User-supplied values still
take priority over these auto-derived defaults via zitadel.mergedConfigmapConfig.
*/}}
{{- define "zitadel.postgresqlAutoConfig" -}}
Database:
  Postgres:
    Host: {{ include "zitadel.postgresqlHost" . }}
    Port: 5432
    Database: {{ .Values.postgresql.auth.database }}
    User:
      Username: {{ .Values.postgresql.auth.username }}
      Password: {{ .Values.postgresql.auth.password }}
      SSL:
        Mode: disable
    Admin:
      Username: postgres
      Password: {{ .Values.postgresql.auth.postgresPassword }}
      SSL:
        Mode: disable
{{- end -}}

{{/*
Returns "configmap" or "dsn" depending on how the user has configured the database.

Detection rule:
- If zitadel.configmapConfig.Database.Postgres.Host is non-empty, the user is in
  "configmap mode" (today's behavior, full backwards compatibility). All discrete
  Database.Postgres.{Host,Port,User,Admin,...} fields are rendered into the
  configmap and ZITADEL reads them from there.
- Otherwise the chart is in "dsn mode": ZITADEL reads its database config from
  the env var ZITADEL_DATABASE_POSTGRES_DSN, supplied either by the user via
  .Values.env / .Values.envVarsSecret, or auto-generated by the chart when the
  bundled postgresql subchart is enabled.

This mirrors how ZITADEL itself decides whether to use the DSN field or the
discrete Host/Port/User/Admin fields on its Postgres config struct (PR #11729).
*/}}
{{- define "zitadel.dbMode" -}}
{{- $host := "" -}}
{{- if (((.Values.zitadel).configmapConfig).Database) -}}
  {{- $host = dig "Postgres" "Host" "" .Values.zitadel.configmapConfig.Database -}}
{{- end -}}
{{- if $host -}}
configmap
{{- else -}}
dsn
{{- end -}}
{{- end -}}

{{/*
Constructs the auto-DSN connection string for the bundled PostgreSQL subchart.
Uses libpq key=value format (not URL form) so we don't have to URL-encode the
password — pgxpool (zitadel) accepts it natively.
The password is referenced via Kubernetes $(VAR) substitution from the
POSTGRES_PASSWORD env var emitted alongside it by zitadel.dbEnv.
*/}}
{{- define "zitadel.bundledPostgresDsn" -}}
host={{ include "zitadel.postgresqlHost" . }} port=5432 user=postgres password=$(POSTGRES_PASSWORD) dbname={{ .Values.postgresql.auth.database }} sslmode=disable
{{- end -}}

{{/*
Returns the env-var entries that should be injected into every container that
talks to the database (main container, init/setup containers).

Always emits ZITADEL_DATABASE_<type>_AWAITINITIALCONN so that ZITADEL
retries its initial database connection on startup instead of crashing
immediately when Postgres is still starting. Users can override via .Values.env.

In bundled-postgres DSN mode this also emits POSTGRES_PASSWORD (sourced from
the Bitnami secret) and ZITADEL_DATABASE_POSTGRES_DSN (referencing it via
$(VAR) substitution). Order matters: POSTGRES_PASSWORD must come first so K8s
can expand it inside the DSN value. .Values.env is appended last so
user-supplied env vars (including a user-provided DSN) always win.

In configmap mode this is just the AwaitInitialConn var plus .Values.env.
*/}}
{{- define "zitadel.dbEnv" -}}
{{- $dbEnv := get (include "zitadel.dbkey.json" . | fromJson) "env" -}}
- name: ZITADEL_DATABASE_{{ $dbEnv }}_AWAITINITIALCONN
  value: "5m"
{{- if and .Values.postgresql.enabled (eq (include "zitadel.dbMode" .) "dsn") }}
- name: POSTGRES_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "zitadel.postgresqlHost" . }}
      key: postgres-password
- name: ZITADEL_DATABASE_POSTGRES_DSN
  value: {{ include "zitadel.bundledPostgresDsn" . | quote }}
{{- end }}
{{- with .Values.env }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{/*
Build the effective configmap config. When postgresql.enabled=true and the user
is in configmap mode, the auto-derived database connection block is merged as a
base and user-supplied configmapConfig values are applied on top (user values
always win).

In DSN mode (the default for the bundled subchart and for any user who doesn't
set Database.Postgres.Host) the chart strips the Database section entirely so
the rendered configmap contains no Database keys; ZITADEL picks up its DB
config from the ZITADEL_DATABASE_POSTGRES_DSN env var instead.
*/}}
{{- define "zitadel.mergedConfigmapConfig" -}}
{{- $config := deepCopy .Values.zitadel.configmapConfig -}}
{{- $mode := include "zitadel.dbMode" . -}}
{{- if eq $mode "dsn" -}}
{{- $_ := unset $config "Database" -}}
{{- else if .Values.postgresql.enabled -}}
{{- $autoConfig := include "zitadel.postgresqlAutoConfig" . | fromYaml -}}
{{- $config = mergeOverwrite $autoConfig $config -}}
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
