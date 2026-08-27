FROM golang:1.26.3-bookworm
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
RUN go build -trimpath -o /usr/local/bin/qecorr ./cmd/qecorr
ENTRYPOINT ["/usr/local/bin/qecorr"]
