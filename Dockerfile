# build image
FROM golang:1.11-alpine as builder
RUN apk update && apk add git 
RUN adduser -D -g '' appuser

COPY . $GOPATH/src/github.com/google-cloud-tools/kube-eagle
WORKDIR $GOPATH/src/github.com/google-cloud-tools/kube-eagle

RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags=”-w -s” -o /go/bin/kube-eagle

# executable image
FROM scratch
COPY --from=builder /go/bin/kube-eagle /go/bin/kube-eagle
USER appuser

ENTRYPOINT ["/go/bin/kube-eagle"]
