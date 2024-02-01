package dag

import (
	"context"
	"os"

	"dagger.io/dagger"
)

func Postgres(client *dagger.Client) *dagger.Service {
	return client.Container().
		From("postgres:alpine").
		WithEnvVariable("POSTGRES_DB", "astro").
		WithEnvVariable("POSTGRES_USER", "postgres").
		WithEnvVariable("POSTGRES_PASSWORD", "postgres").
		WithExposedPort(5432).
		WithExec(nil).
		AsService()
}

func BuildProd() error {
	ctx := context.Background()
	client, err := dagger.Connect(
		ctx,
		dagger.WithWorkdir(".."),
		dagger.WithLogOutput(os.Stdout),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = prodContainer(ctx, client)
	return err
}

func BaseContainer(ctx context.Context, client *dagger.Client) (*dagger.Container, error) {
	dir := client.Host().Directory(".")
	golang := client.Container().
		From("golang:latest").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"mkdir", "/app"}).
		WithWorkdir("/app").
		WithFile("./", dir.File("go.mod")).
		WithMountedFile("./", dir.File("go.sum")).
		WithMountedCache("/root/.cache/go-build", client.CacheVolume("go-build")).
		WithMountedCache("/go/pkg/mod", client.CacheVolume("go-mod")).
		WithMountedDirectory("./", dir)

	return golang, nil
}

func prodContainer(ctx context.Context, client *dagger.Client) (*dagger.Container, error) {
	golang, err := BaseContainer(ctx, client)
	if err != nil {
		return nil, err
	}

	golang = golang.WithExec([]string{"go", "build", "-o", "/go/bin/app", "main.go"})

	distroless := client.Container().
		From("gcr.io/distroless/static:nonroot").
		WithEnvVariable("CONFIG_PATH", "./config.yaml").
		WithMountedFile("/", golang.File("/go/bin/app")).
		WithEntrypoint([]string{"/app"})

	if _, err = distroless.Sync(ctx); err != nil {
		return nil, err
	}

	return distroless, nil
}
