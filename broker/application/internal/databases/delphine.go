package databases

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

type DelphineDb struct {
	Pool *sql.DB
}

func InitialisePool() (*DelphineDb, error) {
	db, err := sql.Open("postgres", os.Getenv("DB_CONNECTION"))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(MustParseInt(os.Getenv("DB_MAX_CONN")))
	db.SetMaxIdleConns(MustParseInt(os.Getenv("DB_MAX_CONN")))
	db.SetConnMaxLifetime(MustParseDurationMinutes(os.Getenv("DB_MAX_CONN_TIME")))
	db.SetConnMaxIdleTime(MustParseDurationMinutes(os.Getenv("DB_MIN_CONN_TIME")))

	// set context background to ping db and asses is all good
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Println("successfully initialised database")
	return &DelphineDb{Pool: db}, nil
}

func MustParseInt(value string) int {
	casted, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		panic(err)
	}
	return int(casted)
}

func MustParseDurationMinutes(value string) time.Duration {
	casted, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}
	return time.Duration(casted) * time.Minute
}
