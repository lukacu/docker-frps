[common]
bind_addr = {{getenv "FRPS_BIND_ADDRESS" "0.0.0.0" }}
bind_port = {{getenv "FRPS_BIND_PORT" "7000" }}

{{if env.Getenv "FRPS_BIND_UDP_PORT" }}
# udp port to help make udp hole to penetrate nat
bind_udp_port = {{getenv "FRPS_BIND_UDP_PORT" }}
{{end}}

{{if env.Getenv "FRPS_KCP_PORT" }}
# udp port used for kcp protocol, it can be same with 'bind_port'
# if not set, kcp is disabled in frps
kcp_bind_port = {{getenv "FRPS_KCP_PORT" }}
{{end}}

# specify which address proxy will listen for, default value is same with bind_addr
# proxy_bind_addr = 127.0.0.1

# if you want to support virtual host, you must set the http port for listening (optional)
# Note: http port and https port can be same with bind_port
vhost_http_port = {{getenv "FRPS_VHOST_HTTP_PORT" "80" }} 
vhost_https_port = {{getenv "FRPS_VHOST_HTTPS_PORT" "443" }}

# response header timeout(seconds) for vhost http server, default is 60s
# vhost_http_timeout = 60

{{if env.Getenv "FRPS_DASHBOARD" }}
dashboard_addr = {{getenv "FRPS_DASHBOARD_ADDRESS" "0.0.0.0" }}
dashboard_port = {{getenv "FRPS_DASHBOARD_PORT" "7500" }}

# dashboard user and passwd for basic auth protect, if not set, both default value is admin
dashboard_user = {{getenv "FRPS_DASHBOARD_USER" "frpsadmin" }}
dashboard_pwd = {{getenv "FRPS_DASHBOARD_PASSWORD" "frpsadmin" }}
{{end}}

{{if env.Getenv "FRPS_LOGFILE" }}
log_file = {{getenv "FRPS_LOGFILE" "/var/log/frps.log" }}

# trace, debug, info, warn, error
log_level = {{getenv "FRPS_LOG_LEVEL" "warn" }}
log_max_days = {{getenv "FRPS_LOG_DAYS" "5" }}
{{end}}

token = {{getenv "FRPS_AUTH_TOKEN" "abcdefghi" }}
allow_ports = 30000-30900

# pool_count in each proxy will change to max_pool_count if they exceed the maximum value
max_pool_count = 5

max_ports_per_client = {{getenv "FRPS_MAX_PORTS" "0" }}

subdomain_host = {{getenv "FRPS_SUBDOMAIN_HOST" "frps.com" }}

tcp_mux = {{getenv "FRPS_TCP_MUX" "true" }}

{{if env.Getenv "FRP_PLUGIN_MULTIUSER" }}
[plugin.multiuser]
addr = 127.0.0.1:9001
path = /handler
ops = Login
{{end}}

{{if env.Getenv "FRP_PLUGIN_ALLOWED_PORTS" }}
[plugin.frp_plugin_allowed_ports]
addr = 127.0.0.1:9002
path = /handler
ops = NewProxy
{{end}}

{{if env.Getenv "FRPS_LINK_NOTIFIER" }}
[plugin.2-linknotifier]
addr = 127.0.0.1:9003
path = /notifier
ops = NewProxy
{{end}}
