###
# Builder container
###
FROM golang:alpine AS builder

ARG tags=none

RUN go version && \
    apk add --update --no-cache gcc musl-dev git curl make gcc g++ python && \
    mkdir /pufferpanel && \
    wget https://github.com/swaggo/swag/releases/download/v1.6.3/swag_1.6.3_Linux_x86_64.tar.gz && \
    tar -zxf swag*.tar.gz && \
    mv swag /go/bin/

WORKDIR /build/apufferi
RUN git clone https://github.com/pufferpanel/apufferi /build/apufferi

WORKDIR /build/pufferd
COPY . .
RUN echo replace github.com/pufferpanel/apufferi/v4 =\> ../apufferi >> go.mod && \
    go get -u github.com/pufferpanel/pufferd/v2/cmd

RUN go build -v -tags $tags -o /pufferpanel/pufferd github.com/pufferpanel/pufferd/v2/cmd

###
# Generate final image
###

FROM alpine
COPY --from=builder /pufferpanel /pufferpanel

EXPOSE 5656 5657
VOLUME /var/lib/pufferd

ENV PUFFERD_CONSOLE_BUFFER=50 \
    PUFFERD_CONSOLE_FORWARD=false \
    PUFFERD_LISTEN_WEB=0.0.0.0:5656 \
    PUFFERD_LISTEN_SOCKET=unix:/var/lib/pufferd/pufferd.sock \
    PUFFERD_LISTEN_WEBCERT=/var/lib/pufferd/web/https.pem \
    PUFFERD_LISTEN_WEBKEY=/var/lib/pufferd/web/https.key \
    PUFFERD_LISTEN_SFTP=0.0.0.0:5657 \
    PUFFERD_LISTEN_SFTPKEY=/var/lib/pufferd/sftp.key \
    PUFFERD_AUTH_PUBLICKEY=/var/lib/pufferd/panel.pem \
    PUFFERD_AUTH_URL=unix:/var/lib/pufferd/pufferpanel.sock \
    PUFFERD_AUTH_CLIENTID=unknown \
    PUFFERD_AUTH_CLIENTSECRET=unknown \
    PUFFERD_DATA_CACHE=/var/lib/pufferd/cache \
    PUFFERD_DATA_SERVERS=/var/lib/pufferd/servers \
    PUFFERD_DATA_MODULES=/var/lib/pufferd/modules \
    PUFFERD_DATA_LOGS=/var/lib/pufferd/logs \
    PUFFERD_DATA_CRASHLIMIT=3

WORKDIR /var/lib/pufferd

ENTRYPOINT ["/pufferpanel/pufferd"]
CMD ["--logging=DEVEL", "run"]