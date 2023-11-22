# build image
FROM docker.io/golang:1.21.4-alpine3.18 as builder
RUN apk update && apk add git ca-certificates

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/kube-eagle

# executable image
FROM scratch
COPY --from=builder /go/bin/kube-eagle /go/bin/kube-eagle

ENV VERSION 1.1.8
ENTRYPOINT ["/go/bin/kube-eagle"]
