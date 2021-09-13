FROM golang:1.16-alpine3.14 as builder
ENV CGO_ENABLED 0

ADD . /src
WORKDIR /src
RUN go get
RUN GOTRACEBACK=all go build -gcflags "all=-N -l" -o metacontroller
# -------------------------
FROM gcr.io/distroless/base
COPY --from=builder /src/metacontroller /metacontroller

USER 10000:10000
EXPOSE 8000

ENTRYPOINT ["/metacontroller"]