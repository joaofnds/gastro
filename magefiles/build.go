//go:build mage

package main

import (
	"context"
	"os"

	"dagger.io/dagger"
)

func Build() error {
	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	dir := client.Host().Workdir()
	golang := client.Container().
		From("golang:1.20").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"mkdir", "/app"}).
		WithWorkdir("/app").
		WithFile("./", dir.File("go.mod")).
		WithFile("./", dir.File("go.sum")).
		WithExec([]string{"go", "mod", "download"}).
		WithMountedDirectory("./", dir).
		WithExec([]string{"go", "build", "-o", "/go/bin/app", "main.go"})

	distroless := client.Container().
		From("gcr.io/distroless/static:nonroot").
		WithEnvVariable("CONFIG_PATH", "./config.yaml").
		WithMountedFile("/", golang.File("/go/bin/app")).
		WithEntrypoint([]string{"/app"})

	_, err = distroless.ExitCode(ctx)

	return err
}
