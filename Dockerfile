FROM golang:alpine AS build
ENV FRP_VER=0.45.0 \
    GOTEMP_VER=3.11.3 \
    FP_MULTIUSER_VER=0.0.2 \
    FRP_ALLOWED_PORTS_VER=1.0.0 \
    OS_NAME=linux \
    OS_TYPE=amd64
    # GO_ACME_LEGO_VER=4.9.0

COPY plugins /src
COPY plugins/frp_allowed_ports/frp_plugin_allowed_ports_v1.0.0_linux_amd64 /usr/local/bin/frp_allowed_ports
RUN apk --no-cache add build-base git gcc && \
    echo "build linknotifier" && cd /src/linknotifier && go mod tidy && go install && go build && \
    echo "download frps" && mkdir /frp && cd /frp && \
    wget https://github.com/fatedier/frp/releases/download/v${FRP_VER}/frp_${FRP_VER}_${OS_NAME}_${OS_TYPE}.tar.gz -O /frp/frp.tar.gz && \
    tar -xvf frp.tar.gz && mv frp_${FRP_VER}_${OS_NAME}_${OS_TYPE}/* ./ && rm /frp/frp.tar.gz && \
    echo "download gomplate" && \
    wget https://github.com/hairyhenderson/gomplate/releases/download/v${GOTEMP_VER}/gomplate_${OS_NAME}-${OS_TYPE}-slim -O /usr/local/bin/gotemp && \
    chmod +x /usr/local/bin/gotemp && \
    echo "download fp-multiuser" && \
    wget https://github.com/gofrp/fp-multiuser/releases/download/v${FP_MULTIUSER_VER}/fp-multiuser-${OS_NAME}-${OS_TYPE} -O /usr/local/bin/fp-multiuser && \
    chmod +x /usr/local/bin/fp-multiuser
    # Using self build bin file due to unable to run git one
    # echo "download frp_plugin_allowed_ports" && \
    # wget https://github.com/Parmicciano/frp_plugin_allowed_ports/releases/download/Release/frp_plugin_allowed_ports -O /usr/local/bin/frp_allowed_ports && \
    # chmod +x /usr/local/bin/frp_allowed_ports
    # TO DO
    # echo "download go-acme lego" && mkdir /lego && cd /lego && \
    # wget https://github.com/go-acme/lego/releases/download/v${GO_ACME_LEGO_VER}/lego_v${GO_ACME_LEGO_VER}_${OS_NAME}_${OS_TYPE}.tar.gz -O /lego/lego.tar.gz && \
    # tar -xvf lego.tar.gz && chmod +x lego && rm /lego/lego.tar.gz

FROM alpine:latest
LABEL MAINTAINER="luka.cehovin@gmail.com kex_zh@outlook.com"
COPY --from=build /frp/frps \
                #   /lego/lego \
                #   /src/portmanager/portmanager \
                  /usr/local/bin/frp_allowed_ports \
                  /src/linknotifier/linknotifier \
                  /usr/local/bin/fp-multiuser \
                  /usr/local/bin/frp_allowed_ports \
                  /usr/local/bin/gotemp \
                  /usr/local/bin/
COPY templates /data
COPY etc /etc
COPY start_runit /sbin/
RUN apk add --no-cache runit
CMD ["/sbin/start_runit"]

# docker build --force-rm -t frps:latest .