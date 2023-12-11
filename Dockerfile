# build image
FROM registry.arthurvardevanyan.com/homelab/toolbox:latest as	builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/kube-eagle

# executable image
FROM scratch
COPY --from=builder /go/bin/kube-eagle /go/bin/kube-eagle

ENV VERSION 1.1.9
ENTRYPOINT ["/go/bin/kube-eagle"]
