package main

import (
	"context"
	"time"

	haikunator "github.com/atrox/haikunatorgo/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	"spacetrouble/internal/pkg/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()
	db, err := pgxpool.Connect(ctx, cfg.DSN())

	if err != nil {
		panic(err)
	}

	const q = "insert into events(id, name, ts) VALUES($1, $2, $3)"

	haikunator := haikunator.New()
	name := haikunator.Haikunate()

	if _, err := db.Exec(ctx, q, uuid.New().String(), name, time.Now().UTC()); err != nil {
		panic(err)
	}
}
