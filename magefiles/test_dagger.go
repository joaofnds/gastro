//go:build mage

package main

import (
	"astro/magefiles/dag"
	"context"
	"os"

	"dagger.io/dagger"
)

func TestDagger() error {
	ctx := context.Background()
	client, err := dagger.Connect(
		ctx,
		dagger.WithWorkdir("."),
		dagger.WithLogOutput(os.Stdout),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	container, err := dag.BaseContainer(ctx, client)
	if err != nil {
		return err
	}

	postgresService := dag.Postgres(client)

	_, err = container.
		WithServiceBinding("postgres", postgresService).
		WithEnvVariable("CONFIG_PATH", "/app/config/config.yaml").
		WithEnvVariable("POSTGRES_HOST", "postgres").
		WithExec([]string{"go", "test", "-v", "./..."}).
		ExitCode(ctx)

	return err
}
