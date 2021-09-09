FROM golang:1.16 as builder
ENV CGO_ENABLED 0

ADD . /src
WORKDIR /src
RUN go get
RUN GOTRACEBACK=all go build -gcflags "all=-N -l" -o metacontroller
#DEBUG:
RUN go get github.com/go-delve/delve/cmd/dlv
# -------------------------
FROM gcr.io/distroless/base
COPY --from=builder /src/metacontroller /metacontroller
#DEBUG:
COPY --from=builder /go/bin/dlv /

USER 10000:10000
EXPOSE 8000
#DEBUG:
EXPOSE 40000

#DEBUG:
ENTRYPOINT [ "./dlv", "--listen=:40000","--headless=true","--api-version=2","--accept-multiclient","exec","./metacontroller" ]
#ENTRYPOINT ["/metacontroller"]