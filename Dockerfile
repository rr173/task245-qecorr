FROM golang:1.26.3-bookworm AS build
WORKDIR /src
COPY . .
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
RUN go build -trimpath -o /out/qecorr ./cmd/qecorr

FROM golang:1.26.3-bookworm
COPY --from=build /out/qecorr /usr/local/bin/qecorr
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/qecorr"]
