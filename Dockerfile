FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./

# Update Tailscale to latest version for compatibility
RUN go get tailscale.com@latest && go mod tidy

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -o railtail -ldflags="-w -s" ./.

FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=builder /app/railtail /usr/local/bin/railtail

ENTRYPOINT ["/usr/local/bin/railtail"]
