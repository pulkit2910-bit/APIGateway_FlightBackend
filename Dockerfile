FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org && go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/server .

CMD ["./server"]