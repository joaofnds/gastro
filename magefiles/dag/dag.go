package dag

import (
	"context"
	"os"

	"dagger.io/dagger"
)

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

	_, err = client.Host().Directory(".").DockerBuild().Sync(ctx)

	return err
}

func BaseContainer(ctx context.Context, client *dagger.Client) (*dagger.Container, error) {
	dir := client.Host().Directory(".")
	golang := client.Container().
		From("golang:1.26.5").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"mkdir", "/app"}).
		WithWorkdir("/app").
		WithMountedCache("/root/.cache/go-build", client.CacheVolume("go-build")).
		WithMountedCache("/go/pkg/mod", client.CacheVolume("go-mod")).
		WithMountedDirectory("./", dir)

	return golang, nil
}
