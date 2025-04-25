FROM golang:1.24 AS builder

ENV CGO_ENABLED=0 

WORKDIR /src
COPY . .

# Remove debug-only resources.
RUN go build \
  -ldflags "-s -w -extldflags '-static'" \ 
  -o /bin/app \
  . \
  && strip /bin/app 

# User info for the scracth
RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc_passwd /etc/passwd
COPY --from=builder --chown=65534:0 /bin/app /app

LABEL org.opencontainers.image.source=https://github.com/onee-only/ff-merge
LABEL org.opencontainers.image.description="Container image used in ff-merge workflow"
LABEL org.opencontainers.image.licenses=MIT

USER nobody
ENTRYPOINT ["/app"]