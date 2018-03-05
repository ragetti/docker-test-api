FROM golang:1.8.3 as builder
WORKDIR /go/src/github.com/ithank
RUN go get -d -v github.com/gorilla/mux
##COPY dockertestapi.go  .
RUN git clone https://github.com/ithank/docker-test-api dockertestapi
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dockertestapi .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/ithank/dockertestapi/dockertestapi .
CMD ["./dockertestapi"]