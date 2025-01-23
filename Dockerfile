FROM golang:1.23-alpine AS go-builder

RUN apk add --no-cache ca-certificates build-base git

WORKDIR /code

ADD go.mod go.sum ./

# Copy over code
COPY . /code

RUN make build \
  && file /code/build/opull

# --------------------------------------------------------
FROM scratch

COPY --from=go-builder /code/build/opull /usr/bin/opull
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt


WORKDIR /opt

ENTRYPOINT ["/usr/bin/opull"]
