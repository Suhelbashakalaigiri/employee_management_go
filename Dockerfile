# ---------- Build Stage ----------
FROM golang:1.27.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o employee-management ./cmd/server


# ---------- Runtime Stage ----------
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/employee-management .

EXPOSE 8081

CMD ["./employee-management"]