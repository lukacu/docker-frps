FROM golang:alpine AS build

RUN apk --no-cache add build-base git gcc

ENV FRP_VERSION 0.34.1

RUN git clone https://github.com/fatedier/frp.git /frp && cd /frp && git reset --hard v${FRP_VERSION}
RUN cd /frp && make

ADD ./plugins /src
RUN cd /src/portmanager && go build
RUN cd /src/acmeproxy && go build
RUN cd /src/linknotifier && go build

FROM alpine:latest
MAINTAINER luka.cehovin@gmail.com

ENV GOTEMP_VERSION 3.5.0

RUN apk add --no-cache wget ca-certificates tar runit

RUN wget https://github.com/hairyhenderson/gomplate/releases/download/v${GOTEMP_VERSION}/gomplate_linux-amd64-slim -O /usr/local/bin/gotemp && \
    chmod +x /usr/local/bin/gotemp

COPY --from=build /frp/bin/frps /usr/local/bin/
COPY --from=build /src/portmanager/portmanager /usr/local/bin/
COPY --from=build /src/acmeproxy/acmeproxy /usr/local/bin/
COPY --from=build /src/linknotifier/linknotifier /usr/local/bin/
COPY start_runit /sbin/
COPY etc /etc/

VOLUME /data

EXPOSE 80 443 7000 7001 7500 30000-30900

CMD ["/sbin/start_runit"]

