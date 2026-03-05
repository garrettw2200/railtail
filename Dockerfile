FROM golang:1.25 AS builder

WORKDIR /app

COPY . ./

# Update Tailscale to latest version for compatibility with p8-cluster
RUN go get tailscale.com@latest && go mod tidy && go mod download

RUN CGO_ENABLED=0 go build -o railtail -ldflags="-w -s" ./.

FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=builder /app/railtail /usr/local/bin/railtail

ENTRYPOINT ["/usr/local/bin/railtail"]
