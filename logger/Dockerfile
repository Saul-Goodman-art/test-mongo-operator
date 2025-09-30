FROM golang:1.23.10 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o logger

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /app/logger .
CMD ["./logger"]