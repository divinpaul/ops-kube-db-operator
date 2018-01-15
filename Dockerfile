FROM myobplatform/go-dep:1.9 as dep

FROM golang:1.9 as builder 
COPY --from=dep /go/bin/dep /go/bin/dep
WORKDIR /go/src/github.com/MYOB-Technology/ops-kube-db-operator
COPY . /go/src/github.com/MYOB-Technology/ops-kube-db-operator
RUN dep ensure -v
RUN go test ./...
ARG VERSION=latest
RUN CGO_ENABLED=0 go build -o /build/postgres-operator -ldflags "-X main.version=$VERSION" -v

FROM scratch
COPY --from=builder /build/postgres-operator /app
ENTRYPOINT [ "/app", "--logtostderr=true"]
