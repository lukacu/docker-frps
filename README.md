Docker container for FRP server.

The image includes two useful server plugins that can be enabled dynamically. The first one is PortManager that maintains a persistent mapping of proxy ports in case the server is restarted. This way the ports will not be redistributed even if the image is upgraded. The second plugin is ACMEProxy, it uses Let's Encrypt (or any other ACME based certificate authority) to automatically secure exposed HTTP connections and redirect them to HTTPS.

Environment configuration:

 * `FRPS_BIND_ADDRESS` - bind to specific address, defaults to 0.0.0.0
 * `FRPS_BIND_PORT` - bind to specific port, default to 7000
 * `FRPS_BIND_UDP_PORT` - UDP port, feature will be disabled if not set
 * `FRPS_KCP_PORT` - KCP port, feature will be disabled if not set
 * `FRPS_VHOST_HTTP_PORT` - vhost http port, default to 80
 * `FRPS_VHOST_HTTPS_PORT` - vhost https port, default to 443
 * `FRPS_DASHBOARD` - set to enable FRPS dashboard
 * `FRPS_DASHBOARD_ADDRESS` - bind dashboard to specific address, defaults to 0.0.0.0
 * `FRPS_DASHBOARD_USER` - username to access dashboard, defaults to "frpsadmin"
 * `FRPS_DASHBOARD_PASSWORD` - password to access dashboard, defaults to "frpsadmin"
 * `FRPS_AUTH_TOKEN` - token for clients, defaults to "abcdefghi"
 * `FRPS_MAX_PORTS` - max ports per client, defaults to unlimited
 * `FRPS_SUBDOMAIN_HOST` - subdomain for virtual hosts, defaults to "frps.com"
 * `FRPS_TCP_MUX` - TCP multiplexing, defaults to true
 * `FRPS_LINK_NOTIFIER` - Enable to turn on LinkNotifier plugin, defaults to false

Note that an external volume has to be mounted to `/data` to make the port reservations and certificates persistent. 

### LinkNotifier

Plugin can notify user of its active/inactive proxy ports via email. The following information must be provided when starting docker:

 * `FRPS_LINK_NOTIFIER` - set environment var to enable the plugin
 * `FRPS_LINK_NOTIFIER_SMTP_SERVER`- set environment var to SMTP server in format `hostname.com:port` 
 * `FRPS_LINK_NOTIFIER_SMTP_ACCOUNT` - set environment var to your email/account name
 * `FRPS_LINK_NOTIFIER_SMTP_PASS` - set environment var to your password 
 * `FRPS_LINK_NOTIFIER_EMAIL_SUBJECT` - set environment var to subject of the email, defaults to "Reverse proxy links update"
 * `FRPS_LINK_NOTIFIER_DELAY_SEC`- set environment var to seconds of delay after last modification has been done before sending notification, defaults to 15
 * `FRPS_LINK_NOTIFIER_SLEEP_CHECK_SEC`- set environment var to seconds of sleep time in infinite loop, defaults to 5
 * `FRPS_LINK_NOTIFIER_CONNECTION_CHECK_TIMEOUT_SEC` - set environment var to second of timeout after port connection is considered inactive, default to 2

Template of the email must be provided in `/data/notification_email.html.tpl`. Template is run for each email notification and passes `DisplayProxyList` struct, which groups proxies of the same user. `Active` and `Inactive` group proxies with the same ContainerName for the given user.

```go

type DisplayProxyList struct {
    Active      map[string][]ProxyInfo    `json:"active"`
    Inactive    map[string][]ProxyInfo    `json:"inactive"`
}

type ProxyInfo struct {
    Name          string     `json:"name"`
    ContainerName string     `json:"container_name"`
    ProxyType     string     `json:"proxy_type"`
    RemotePort    int        `json:"remote_port"`
    LocalPort     int        `json:"local_port"`
    Email         string     `json:"email"`
    ClientPrefix  string     `json:"frps_prefix"`
    Url           string     `json:"url"`
    Active        bool       `json:"active"`
    Notified      bool       `json:"notified"`   
}
```

To activate the email notification, FRP client must provide the following meta data in its configuration for each proxy connection:
 * `meta_notify_email` - set to email address that will recieve the notificaiton
 * `meta_local_port` - set to local port used (the same as local_port)
 * `meta_frpc_prefix` - set to FRP client specific name (e.g., server hostname)

### Reference
 * [fp-multiuser](https://github.com/gofrp/fp-multiuser)
 * [frp_plugin_allowed_ports](https://github.com/Parmicciano/frp_plugin_allowed_ports)
