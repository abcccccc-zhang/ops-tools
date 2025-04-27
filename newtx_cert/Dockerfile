FROM golang:1.22.2 AS builder

WORKDIR /app
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
        GOPROXY="https://goproxy.cn,direct"

COPY go.mod go.sum ./

RUN go mod download


COPY . .

RUN go build -o newcert .
FROM alpine:latest
ENV TZ=Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
COPY --from=builder /app/newcert /newcert
CMD ["/newcert"]

