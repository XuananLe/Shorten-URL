package stores

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"shorten-url/backend/pkg/db/sqlc"
	"shorten-url/backend/pkg/utils"
)

type Postgres struct {
	DB      *pgxpool.Pool
	Queries *sqlc.Queries
}

var PostgresClient *Postgres = &Postgres{}

func InitPostgres() *Postgres {
	dbUrl := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", utils.Config.Database.DB_USERNAME, utils.Config.Database.DB_PASSWORD, utils.Config.Database.DB_PORT, utils.Config.Database.DB_NAME)
	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	PostgresClient.Queries = sqlc.New(pool)
	PostgresClient.DB = pool
	poolConfig, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		log.Fatal("Unable to parse config:", err)
	}
	poolConfig.MaxConns = 100
	PostgresClient.DB, _ = pgxpool.NewWithConfig(context.Background(), poolConfig)
	
	err = PostgresClient.DB.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return PostgresClient
}
