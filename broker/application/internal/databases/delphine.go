package databases

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	Database *DelphineDatabase
	once     sync.Once
)

type DelphineDatabase struct {
	Pool *pgxpool.Pool
}

func InitDelphineDatabase() *DelphineDatabase {
	once.Do(func() {
		log.Print("initialise database")
		poolConfig, err := pgxpool.ParseConfig(os.Getenv("DB_CONNECTION"))
		if err != nil {
			log.Fatalf("%s,Unable to parse connection config: %v", Now(), err)
		}

		poolConfig.MaxConns = MustParseInt32(os.Getenv("DB_MAX_CONN"))                       // Maximum number of connections in the pool
		poolConfig.MinConns = MustParseInt32(os.Getenv("DB_MIN_CONN"))                       // Minimum number of connections to maintain
		poolConfig.MaxConnLifetime = MustParseDurationMinutes(os.Getenv("DB_MAX_CONN_TIME")) // Maximum lifetime of a connection
		poolConfig.MaxConnIdleTime = MustParseDurationMinutes(os.Getenv("DB_MIN_CONN_TIME")) // Maximum idle time for connections

		pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			log.Fatalf("%s,Unable to create connection pool: %v", Now(), err)
		}
		Database = &DelphineDatabase{
			Pool: pool,
		}
	})

	return Database
}

func GetDatabase() *DelphineDatabase {
	return Database
}

func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func MustParseInt32(value string) int32 {
	casted, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		panic(err)
	}
	return int32(casted)
}

func MustParseDurationMinutes(value string) time.Duration {
	casted, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}
	return time.Duration(casted) * time.Minute
}
