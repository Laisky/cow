FROM golang:1.19.2-bullseye AS builder

# install dependencies
RUN apt-get update \
    && apt-get install -y --no-install-recommends g++ make gcc git build-essential ca-certificates curl \
    && update-ca-certificates

ENV GO111MODULE=on
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

# static build
ADD . .
RUN go build -a -ldflags '-w -extldflags "-static"' -o main .


# copy executable file and certs to a pure container
FROM debian:11.4

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates haveged \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/main /app/cow
COPY --from=builder /app/doc/sample-config/rc-en /etc/cow/rc

WORKDIR /app
RUN chmod +rx -R /app && \
    adduser --disabled-password --gecos '' laisky
USER laisky

ENTRYPOINT [ "/app/cow" ]
CMD [ "-rc", "/etc/cow/rc" ]
