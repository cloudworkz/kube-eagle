# build image
FROM golang:1.11-alpine as builder
RUN apk update && apk add git

COPY . $GOPATH/src/github.com/google-cloud-tools/kube-eagle/
WORKDIR $GOPATH/src/github.com/google-cloud-tools/kube-eagle/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/kube-eagle

# executable image
FROM scratch
COPY --from=builder /go/bin/kube-eagle /go/bin/kube-eagle

ENTRYPOINT ["/go/bin/kube-eagle"]
