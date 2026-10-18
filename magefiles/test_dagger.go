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
		dagger.WithWorkdir(".."),
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

	_, err = container.
		WithEnvVariable("CONFIG_PATH", "/app/config/config.yaml").
		WithExec([]string{"go", "test", "-v", "./..."}).
		Sync(ctx)

	return err
}
