package pkg

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func NewServer(ctx context.Context, configFile, envPrefix string) (*http.Server, error) {
	cfg, err := loadConfig(configFile, envPrefix)
	if err != nil {
		return nil, err
	}

	r := router(ctx, cfg)
	workDir, _ := os.Getwd()
	dist := http.Dir(filepath.Join(workDir, "web", "dist"))
	fileServer(r, "/static", dist)

	return &http.Server{
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSecs) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSecs) * time.Second,
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      r,
	}, nil
}
