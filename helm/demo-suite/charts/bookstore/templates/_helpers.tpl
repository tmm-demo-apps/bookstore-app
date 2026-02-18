{{/*
Bookstore fullname
*/}}
{{- define "bookstore.fullname" -}}
{{- printf "%s-bookstore" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bookstore.labels" -}}
app.kubernetes.io/name: bookstore
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels for app
*/}}
{{- define "bookstore.appSelectorLabels" -}}
app: bookstore-app
{{- end }}

{{/*
Resolve ingress host: explicit .Values.ingress.host > auto from global.domain
*/}}
{{- define "bookstore.ingressHost" -}}
{{- if .Values.ingress.host -}}
  {{- .Values.ingress.host -}}
{{- else -}}
  {{- printf "bookstore.%s" .Values.global.domain -}}
{{- end -}}
{{- end }}

{{/*
ImagePullSecrets from global
*/}}
{{- define "bookstore.imagePullSecrets" -}}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
{{- range .Values.global.imagePullSecrets }}
  - name: {{ .name }}
{{- end }}
{{- end }}
{{- end }}

{{/*
StorageClassName - empty string means cluster default
*/}}
{{- define "bookstore.storageClassName" -}}
{{- if .Values.global.storageClassName -}}
storageClassName: {{ .Values.global.storageClassName }}
{{- end -}}
{{- end }}
