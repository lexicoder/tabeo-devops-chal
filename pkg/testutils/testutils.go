// +build test

package testutils

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func GetTestDb() (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("host=localhost port=%s dbname=space user=space password=secure pool_max_conns=99", os.Getenv("DB_TEST_PORT"))
	ctx := context.Background()
	db, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return db, err
	}
	if err := db.Ping(ctx); err != nil {
		return db, err
	}
	return db, err
}

func SpinPostgresContainer(ctx context.Context, rootDir string) testcontainers.Container {
	mountFrom := mergeMigrations(rootDir)
	defer func() {
		if strings.HasSuffix(mountFrom, "test-db.init.sql") {
			os.Remove(mountFrom)
		}
	}()
	mountTo := "/docker-entrypoint-initdb.d/init.sql"

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{"5432/tcp"},
		BindMounts:   map[string]string{mountFrom: mountTo},
		Env: map[string]string{
			"POSTGRES_DB":       "space",
			"POSTGRES_USER":     "space",
			"POSTGRES_PASSWORD": "secure",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	postgresContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true},
	)
	if err != nil {
		panic(err)
	}
	p, _ := postgresContainer.MappedPort(ctx, "5432")
	os.Setenv("DB_TEST_PORT", p.Port())
	// added this sleep here :(
	// better to find a way to proper wait
	time.Sleep(5 * time.Second)
	return postgresContainer
}

func mergeMigrations(root string) string {
	var alllines []string
	err := filepath.Walk(root+"migrations/", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		lines, err := readLines(path)
		if err != nil {
			return err
		}
		for _, l := range lines {
			if !strings.HasPrefix(l, "----") {
				alllines = append(alllines, l)
			} else {
				break
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	resultPath := root + "test-db.init.sql"
	if err := writeLines(alllines, resultPath); err != nil {
		panic(err)
	}
	return resultPath
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
