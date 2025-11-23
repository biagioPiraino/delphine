package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/biagioPiraino/delphico/consumer/internal/databases"
	"github.com/google/uuid"
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

type FinanceRepository struct {
	db *databases.DelphineDb
}

func NewFinanceRepository(db *databases.DelphineDb) *FinanceRepository {
	return &FinanceRepository{db: db}
}

func (r FinanceRepository) AddFinanceMetadata(metadata FinanceArticleMetadata) error {
	requestId := uuid.New()
	tx, err := r.db.Pool.Begin() // begin transaction
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx, requestId uuid.UUID) {
		if err := tx.Rollback(); err != nil { // safety if commit never reached
			log.Println(fmt.Sprintf("rollback for request %s not applied", requestId))
		} else {
			log.Println(fmt.Sprintf("rollback for request %s applied succesfully", requestId))
		}
	}(tx, requestId)

	procedure := `CALL kb_finance.add_metadata(
		ROW($1,$2,$3,$4,$5)::kb_finance.article_metadata_type
	)`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = tx.ExecContext(ctx, procedure,
		requestId,
		metadata.Author,
		metadata.Provider,
		metadata.Url,
		metadata.Published,
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	} else {
		log.Println(fmt.Sprintf("request %s committed", requestId))
	}

	return nil
}

func (r FinanceRepository) AddFinanceArticle(article FinanceArticle) error {
	requestId := uuid.New()
	tx, err := r.db.Pool.Begin()
	if err != nil {
		return err
	}

	defer func(tx *sql.Tx, requestId uuid.UUID) {
		if err := tx.Rollback(); err != nil { // safety if commit never reached
			log.Println(fmt.Sprintf("rollback for request %s not applied", requestId))
		} else {
			log.Println(fmt.Sprintf("rollback for request %s applied succesfully", requestId))
		}
	}(tx, requestId)

	procedure := `CALL kb_finance.add_article(
		ROW($1,$2,$3)::kb_finance.article_type
	)`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = tx.ExecContext(ctx, procedure,
		uuid.New(),
		article.Url,
		article.Content,
	)

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	} else {
		log.Println(fmt.Sprintf("request %s committed", requestId))
	}

	return nil
}
