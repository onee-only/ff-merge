FROM golang:1.24 AS builder

ENV CGO_ENABLED=0 

# Download the compressor
RUN apt-get -qq update && \
  apt-get -yqq install upx

WORKDIR /src
COPY . .

# Remove debug-only resources and compress the binary.
RUN go build \
  -ldflags "-s -w -extldflags '-static'" \ 
  -o /bin/app \
  . \
  && strip /bin/app \
  && upx -q -9 /bin/app

# User info for the scracth
RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc_passwd /etc/passwd
COPY --from=builder --chown=65534:0 /bin/app /app

USER nobody
ENTRYPOINT ["/app"]