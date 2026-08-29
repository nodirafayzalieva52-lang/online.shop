FROM golang:1.24-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o shop ./main.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /app/shop /app/shop

EXPOSE 8080

CMD ["/app/shop"]
