FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/auditor-server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/auditor-server /auditor-server
VOLUME ["/data"]
ENV LISTEN_ADDR=:8080 LEDGER_PATH=/data/ledger.json
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/auditor-server"]
