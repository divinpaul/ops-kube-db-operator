FROM golang:1.9 as builder

RUN curl https://glide.sh/get | sh

WORKDIR /go/src/github.com/MYOB-Technology/ops-kube-db-operator
COPY . /go/src/github.com/MYOB-Technology/ops-kube-db-operator
RUN glide install

ARG VERSION=latest
RUN CGO_ENABLED=0 go build -o /build/postgres-operator -ldflags "-X main.version=$VERSION" -v

FROM scratch
COPY --from=builder /build/postgres-operator /app
ENTRYPOINT [ "/app", "--logtostderr=true"]
