package migrations

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var migrationFiles embed.FS

func Run(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := migrationFiles.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	var files []string
	for _, entry := range entries {
		// Выполняем только накатывающие миграции (.up.sql)
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		content, err := migrationFiles.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file, err)
		}

		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			log.Printf("migration %s skipped or failed: %v", file, err)
		} else {
			log.Printf("migration applied successfully: %s", file)
		}
	}

	return nil
}