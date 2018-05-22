FROM golang:1.9.0 AS builder
WORKDIR /go/src/github.com/szymon-planeta/predicted_cpu_exporter
COPY . .
RUN go get -d
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -tags netgo -ldflags '-w'

# Final image.
FROM scratch
LABEL maintainer "Szymon Planeta <planetaszymon@gmail.com>"
COPY --from=builder /go/src/github.com/szymon-planeta/predicted_cpu_exporter .
EXPOSE 8080
ENTRYPOINT ["/predicted_cpu_exporter"]
