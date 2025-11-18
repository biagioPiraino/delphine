package repositories

import (
	"context"
	"errors"

	"github.com/biagioPiraino/delphico/consumer/internal/databases"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FinanceArticleMetadata struct {
	Url       string `json:"url"`
	Author    string `json:"author"`
	Published string `json:"published"`
	Provider  string `json:"provider"`
}

type FinanceArticle struct {
	Url     string `json:"url"`
	Content string `json:"content"`
}

func AddFinanceMetadata(metadata FinanceArticleMetadata) error {
	connection := acquireConnection()
	if connection == nil {
		return errors.New("database not initiated")
	}
	defer connection.Release()

	procedure := `CALL kb_finance.add_metadata(
		ROW($1,$2,$3,$4,$5)::kb_finance.article_metadata_type
	)`

	_, err := connection.Exec(context.Background(), procedure,
		uuid.New(),
		metadata.Author,
		metadata.Provider,
		metadata.Url,
		metadata.Published,
	)

	return err
}

func AddFinanceArticle(article FinanceArticle) error {
	connection := acquireConnection()
	if connection == nil {
		return errors.New("database not initiated")
	}
	defer connection.Release()

	procedure := `CALL kb_finance.add_article(
		ROW($1,$2,$3)::kb_finance.article_type
	)`

	_, err := connection.Exec(context.Background(), procedure,
		uuid.New(),
		article.Url,
		article.Content,
	)

	return err
}

func acquireConnection() *pgxpool.Conn {
	db := databases.GetDatabase()
	// database not instantiated
	if db == nil {
		return nil
	}

	// failing acquiring connection from pool
	connection, err := db.Pool.Acquire(context.Background())
	if err != nil {
		return nil
	}

	return connection
}
