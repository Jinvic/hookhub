# syntax=docker/dockerfile:1

FROM golang:1.25-bookworm AS build
WORKDIR /src

# go.mod may require a newer toolchain; let the Go wrapper fetch it.
ENV GOTOOLCHAIN=auto
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/hookhub ./cmd/hookhub

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/hookhub /app/hookhub
COPY config.example.yaml /app/config.example.yaml

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/hookhub"]
CMD ["-config", "/app/config.yaml"]
