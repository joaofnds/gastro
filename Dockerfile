FROM --platform=$BUILDPLATFORM golang:1.26.5 AS build
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o /out/astro ./cmd/astro \
  && go build -o /out/migrate ./cmd/migrate \
  && go build -o /out/transfer ./cmd/transfer

FROM litestream/litestream:0.5.17@sha256:4b02b9859a6b6b4087d8b8944e15f7e984bd7957cba322bbeee38b0e27b9656a
ENV SQLITE_PATH=/data/astro.db
COPY --from=build /out/astro /out/migrate /out/transfer /
COPY ops/litestream.yml /etc/litestream.yml
COPY --chmod=755 scripts/entrypoint.sh scripts/serve.sh scripts/recover.sh /scripts/
USER 65532:65532
ENTRYPOINT ["/scripts/entrypoint.sh"]
CMD ["serve"]
