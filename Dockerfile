# Build stage
FROM golang:1.24.2-alpine AS builder

RUN apk --no-cache add build-base git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o pscloud-exporter ./cmd/pscloud-exporter

FROM scratch

COPY --from=builder /app/pscloud-exporter /pscloud-exporter

ENTRYPOINT ["/pscloud-exporter"]
CMD ["--version"]