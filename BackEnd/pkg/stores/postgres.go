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
	dbUrl := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", utils.Config.Database.DbUsername,
		utils.Config.Database.DbPassword, utils.Config.Database.DbPort, utils.Config.Database.DbName)
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
	poolConfig.MaxConns = 10000
	PostgresClient.DB, _ = pgxpool.NewWithConfig(context.Background(), poolConfig)

	err = PostgresClient.DB.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Postgres Connected")
	return PostgresClient
}
