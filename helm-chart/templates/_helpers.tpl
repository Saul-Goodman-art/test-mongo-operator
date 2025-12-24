{{- define "orders-app.fullname" -}}
{{- printf "%s" .Release.Name -}}
{{- end -}}
{{- define "orders-app.mongodbServiceName" -}}
{{- if .Values.mongodb.serviceName -}}
{{- .Values.mongodb.serviceName -}}
{{- else -}}
{{- include "orders-app.fullname" . }}-mongodb-svc
{{- end -}}
{{- end -}}