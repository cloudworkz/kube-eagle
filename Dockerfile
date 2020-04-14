# build image
FROM golang:1.13-alpine as builder
RUN apk update && apk add git ca-certificates

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/kube-eagle

# executable image
FROM scratch
COPY --from=builder /go/bin/kube-eagle /go/bin/kube-eagle

EXPOSE ${TELEMETRY_PORT}

ENV VERSION 1.1.3
ENTRYPOINT ["/go/bin/kube-eagle"]
