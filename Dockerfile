FROM golang:1.22-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY ./ /app/
RUN go build -ldflags="-s -w" -o /app/main cmd/server/main.go


FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app /app

EXPOSE 9898

CMD ["./main"]
