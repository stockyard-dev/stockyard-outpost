FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go mod download && CGO_ENABLED=0 go build -o outpost ./cmd/outpost/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/outpost .
ENV PORT=9220 DATA_DIR=/data
EXPOSE 9220
CMD ["./outpost"]
