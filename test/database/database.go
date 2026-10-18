package database

import (
	"astro/adapters/sqlite"
	"context"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"go.uber.org/fx"
)

var Module = fx.Decorate(isolatedConfig)

func isolatedConfig(_ sqlite.Config) (sqlite.Config, error) {
	config := sqlite.Config{Path: filepath.Join(ginkgo.GinkgoT().TempDir(), "astro.db")}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return config, sqlite.Initialize(ctx, config.Path)
}
