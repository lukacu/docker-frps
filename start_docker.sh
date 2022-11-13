docker run --rm -e FRPS_DASHBOARD=true \
                 -e FRP_PLUGIN_MULTIUSER=true \
                 -e FRP_PLUGIN_ALLOWED_PORTS=true \
                 -p 7000:7000 -p 7500:7500 \
                 -v "$(pwd)/data/tokens:/data/tokens" \
                 -v "$(pwd)/data/ports:/data/ports" \
                 frps:latest

# docker run --rm -e FRPS_DASHBOARD=true \
#                  -e FRP_PLUGIN_MULTIUSER=true \
#                  -e FRP_PLUGIN_ALLOWED_PORTS=true \
#                  -p 7000:7000 -p 7500:7500 \
#                  frps:latest

# docker run --rm -it -e FRPS_DASHBOARD=true \
#                  -p 7000:7000 -p 7500:7500 \
#                  frps:latest /bin/sh