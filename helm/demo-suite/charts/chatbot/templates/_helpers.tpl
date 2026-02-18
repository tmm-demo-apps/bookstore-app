{{/*
Chatbot fullname
*/}}
{{- define "chatbot.fullname" -}}
{{- printf "%s-chatbot" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "chatbot.labels" -}}
app.kubernetes.io/name: chatbot
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "chatbot.selectorLabels" -}}
app: chatbot
{{- end }}

{{/*
Resolve ingress host
*/}}
{{- define "chatbot.ingressHost" -}}
{{- if .Values.ingress.host -}}
  {{- .Values.ingress.host -}}
{{- else -}}
  {{- printf "chatbot.%s" .Values.global.domain -}}
{{- end -}}
{{- end }}

{{/*
ImagePullSecrets from global
*/}}
{{- define "chatbot.imagePullSecrets" -}}
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
{{- define "chatbot.storageClassName" -}}
{{- if .Values.global.storageClassName -}}
storageClassName: {{ .Values.global.storageClassName }}
{{- end -}}
{{- end }}
