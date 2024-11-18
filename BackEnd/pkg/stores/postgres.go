package stores

import (
	"context"
	"fmt"
	"shorten-url/backend/pkg/config"
	"shorten-url/backend/pkg/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

type Postgres struct {
	DB      *pgxpool.Pool
	Queries *sqlc.Queries
}

var PostgresClient *Postgres = &Postgres{}

func InitPostgres() *Postgres {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", 
	config.AppConfig.Database.Username,
		config.AppConfig.Database.Password,
		config.AppConfig.Database.Host,
		config.AppConfig.Database.Port, 
		config.AppConfig.Database.Name)

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
	poolConfig.MaxConns = 1000
	PostgresClient.DB, _ = pgxpool.NewWithConfig(context.Background(), poolConfig)

	err = PostgresClient.DB.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Postgres Connected")
	return PostgresClient
}
