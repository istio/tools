# Note: this image must be built from the root of the repository for access to
# the vendor folder.

FROM golang:1.10.2 AS builder

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/istio.io/tools

COPY . .
RUN dep ensure -vendor-only

WORKDIR /go/src/istio.io/tools/isotope/service

RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo -o ./main ./main.go

FROM scratch
COPY --from=builder \
    /go/src/istio.io/tools/isotope/service/main /usr/local/bin/service

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/service"]
