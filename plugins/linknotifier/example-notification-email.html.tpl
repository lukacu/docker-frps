This is an automated message from auto-notifier that manages reverse proxy connections to docker containers. URLs and ports to containers may have been updated. 

You now have the following ACTIVE containers:
{{ range $n, $p_list := .Active }}
{{$n -}}:
{{- range $key, $p:= $p_list }}
  {{$p.Url}} -> local port {{$p.LocalPort}} (on GPU server '{{ $p.ClientPrefix }}') {{ if not $p.Notified }}**NEW**{{end}}
{{- end }}
{{- end }}

The following connections to containers are NOT active:
{{ range $n, $p_list := .Inactive }}
{{$n -}}:
{{- range $key, $p:= $p_list }}
  {{$p.Url}} -> local port {{$p.LocalPort}} (on GPU server '{{ $p.ClientPrefix }}')
{{- end }}
{{- end }}

Non-active container connections may be due to stopped/removed container, or container is started but the service at the local port inside the container is not responsive.

For any question, please email cluster maintainers: custer-manager@mydomain.com

