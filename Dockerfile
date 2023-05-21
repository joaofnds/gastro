FROM golang:1.20 as build
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . /app
RUN go build -o /go/bin/app cmd/astro/astro.go

FROM gcr.io/distroless/static:nonroot
COPY --from=build /go/bin/app /
CMD ["/app"]
