{{- $db := .inputs.DB -}}
{{- $chComponents := list -}}
{{- range .inputs.DB.Spec.Components -}}
  {{- if eq .Type "clickhouse" -}}
    {{- $chComponents = append $chComponents . -}}
  {{- end -}}
{{- end -}}
{{- if eq (len $chComponents) 1 -}}
{{- $component := index $chComponents 0 -}}
apiVersion: clickhouse.altinity.com/v1
kind: ClickHouseInstallation
metadata:
  name: {{ $db.Name }}
  namespace: {{ $db.Namespace }}
spec:
  configuration:
    users:
      admin/password:
        valueFrom:
          secretKeyRef:
            name: {{ $db.Name }}-admin-password
            key: password
    {{- if $component.CustomSpec }}
    {{- $customSpec := fromJson (toString $component.CustomSpec) }}
    {{- if $customSpec.zookeeper }}
    zookeeper:
      {{- toYaml $customSpec.zookeeper | nindent 6 }}
    {{- end }}
    {{- end }}
    clusters:
    - name: {{ $component.Name }}
      {{- if or $component.Shards $component.Replicas }}
      layout:
        {{- if $component.Shards }}
        shardsCount: {{ $component.Shards }}
        {{- end }}
        {{- if $component.Replicas }}
        replicasCount: {{ $component.Replicas }}
        {{- end }}
      {{- end }}
      templates:
        podTemplate: clickhouse-default
  templates:
    podTemplates:
    - name: clickhouse-default
      metadata:
        {{- if $component.PodSpec.Labels }}
        labels:
          {{- toYaml $component.PodSpec.Labels | nindent 10 }}
        {{- end }}
        {{- if $component.PodSpec.Annotations }}
        annotations:
          {{- toYaml $component.PodSpec.Annotations | nindent 10 }}
        {{- end }}
      spec:
        containers:
        - name: {{ if $component.PodSpec.Container.Name }}{{ $component.PodSpec.Container.Name }}{{ else }}clickhouse{{ end }}
          {{- if $component.Image }}
          image: {{ $component.Image }}
          {{- else if $component.PodSpec.Container.Image }}
          image: {{ $component.PodSpec.Container.Image }}
          {{- end }}
          {{- if $component.PodSpec.Container.Resources }}
          resources:
            {{- toYaml $component.PodSpec.Container.Resources | nindent 12 }}
          {{- end }}
          {{- if $component.PodSpec.Container.Env }}
          env:
            {{- toYaml $component.PodSpec.Container.Env | nindent 12 }}
          {{- end }}
          volumeMounts:
          - name: data
            mountPath: /var/lib/clickhouse
          {{- if $component.PodSpec.Container.VolumeMounts }}
          {{- toYaml $component.PodSpec.Container.VolumeMounts | nindent 10 }}
          {{- end }}
        {{- if $component.PodSpec.Sidecars }}
        {{- toYaml $component.PodSpec.Sidecars | nindent 8 }}
        {{- end }}
    volumeClaimTemplates:
    - name: data
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: {{ $component.Storage.Size }}
        {{- if $component.Storage.StorageClass }}
        storageClassName: {{ $component.Storage.StorageClass }}
        {{- end }}
    {{- if $component.PodSpec.AdditionalVolumeClaimTemplates }}
    {{- range $component.PodSpec.AdditionalVolumeClaimTemplates }}
    - name: {{ .Name }}
      spec:
        {{- toYaml .Spec | nindent 8 }}
    {{- end }}
    {{- end }}
{{- else }}
{{- fail "Expected exactly one clickhouse component" }}
{{- end }}