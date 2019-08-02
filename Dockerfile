FROM alpine:latest
MAINTAINER luka.cehovin@gmail.com

ENV FRP_VERSION 0.28.0
ENV GOTEMP_VERSION 3.5.0

RUN apk add --no-cache wget ca-certificates tar runit

RUN mkdir /frp/ && cd /frp && \
    wget https://github.com/fatedier/frp/releases/download/v${FRP_VERSION}/frp_${FRP_VERSION}_linux_amd64.tar.gz -O frp.tar.gz && \
    tar xvzf frp.tar.gz && mv frp_${FRP_VERSION}_linux_amd64/frps /usr/local/bin/ && cd / && rm -rf /frp && \
    wget https://github.com/hairyhenderson/gomplate/releases/download/v${GOTEMP_VERSION}/gomplate_linux-amd64-slim -O /usr/local/bin/gotemp && \
    chmod +x /usr/local/bin/gotemp

COPY start_runit /sbin/
COPY etc /etc/

EXPOSE 80 443 7000 7001 7500 30000-30900

CMD ["/sbin/start_runit"]
