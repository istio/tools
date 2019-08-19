# Note: this image must be built from the root of the repository for access to
# the vendor folder.

FROM golang:1.12.7 AS builder

WORKDIR /go/src/istio.io/tools

COPY . .

WORKDIR /go/src/istio.io/tools/isotope/service

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo -o ./main ./main.go

FROM scratch
COPY --from=builder \
    /go/src/istio.io/tools/isotope/service/main /usr/local/bin/isotope_service

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/isotope_service"]
