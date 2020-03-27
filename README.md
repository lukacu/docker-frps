Docker container for FRP server.

The image includes two useful server plugins that can be enabled dynamically. The first one is PortManager that maintains a persistent mapping of proxy ports in case the server is restarted. This way the ports will not be redistributed even if the image is upgraded. The second plugin is ACMEProxy, it uses Let's Encrypt (or any other ACME based certificate authority) to automatically secure exposed HTTP connections and redirect them to HTTPS.

Environment configuration:

 * `FRPS_BIND_ADDRESS` - bind to specific address, defaults to 0.0.0.0
 * `FRPS_DASHBOARD` - set to enable FRPS dashboard
 * `FRPS_DASHBOARD_ADDRESS` - bind dashboard to specific address, defaults to 0.0.0.0
 * `FRPS_DASHBOARD_USER` - username to access dashboard, defaults to "frpsadmin"
 * `FRPS_DASHBOARD_PASSWORD` - password to access dashboard, defaults to "frpsadmin"
 * `FRPS_AUTH_TOKEN` - token for clients, defaults to "abcdefghi"
 * `FRPS_MAX_PORTS` - max ports per client, defaults to unlimited
 * `FRPS_SUBDOMAIN_HOST` - subdomain for virtual hosts, defaults to "frps.com"
 * `FRPS_TCP_MUX` - TCP multiplexing, defaults to true
 * `FRPS_PERSISTENT_PORTS` - Enable to turn on PortManager plugin, defaults to false
 * `FRPS_LETSENCRYPT_EMAIL` - Set to your email to enable ACMEProxy, defaults to empty string

Note that an external volume has to be mounted to `/data` to make the port reservations and certificates persistent.
