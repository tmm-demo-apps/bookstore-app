{{/*
Reader fullname
*/}}
{{- define "reader.fullname" -}}
{{- printf "%s-reader" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "reader.labels" -}}
app.kubernetes.io/name: reader
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "reader.selectorLabels" -}}
app: reader
{{- end }}

{{/*
Resolve ingress host
*/}}
{{- define "reader.ingressHost" -}}
{{- if .Values.ingress.host -}}
  {{- .Values.ingress.host -}}
{{- else -}}
  {{- printf "reader.%s" .Values.global.domain -}}
{{- end -}}
{{- end }}

{{/*
ImagePullSecrets from global
*/}}
{{- define "reader.imagePullSecrets" -}}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
{{- range .Values.global.imagePullSecrets }}
  - name: {{ .name }}
{{- end }}
{{- end }}
{{- end }}

{{/*
StorageClassName
*/}}
{{- define "reader.storageClassName" -}}
{{- if .Values.global.storageClassName -}}
storageClassName: {{ .Values.global.storageClassName }}
{{- end -}}
{{- end }}
