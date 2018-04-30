FROM golang:1.8.3 as builder
WORKDIR /go/src/dockertestapi

COPY . .
#RUN git clone https://github.com/ithank/docker-test-api dockertestapi
#RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dockertestapi .

RUN go get -d -v ./...
RUN go install -v ./...
RUN CGO_ENABLED=0 go build -o /app/dockertestapi

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app .

EXPOSE 8000
CMD ["/app/dockertestapi"]
