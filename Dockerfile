FROM golang:1.21 as build
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build cmd/astro/astro.go \
  && go build cmd/migrate/migrate.go

FROM gcr.io/distroless/static:nonroot
COPY --from=build /app/astro /app/migrate /
CMD ["/astro"]
