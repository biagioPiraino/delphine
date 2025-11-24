package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/biagioPiraino/delphico/consumer/internal/databases"
	"github.com/google/uuid"
)

type Article struct {
	Url       string `json:"url"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Published string `json:"published"`
	Provider  string `json:"provider"`
	Content   string `json:"content"`
	Domain    string `json:"domain"`
}

type FinanceRepository struct {
	db *databases.DelphineDb
}

func NewFinanceRepository(db *databases.DelphineDb) *FinanceRepository {
	return &FinanceRepository{db: db}
}

func (r FinanceRepository) AddFinanceArticle(article Article) error {
	requestId := uuid.New()
	tx, err := r.db.Pool.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(); err != nil { // safety if commit never reached
			fmt.Println(fmt.Sprintf("rollback for request %s not applied", requestId))
		} else {
			fmt.Println(fmt.Sprintf("rollback for request %s applied succesfully", requestId))
		}
	}()

	procedure := `CALL kb.add_article(
		ROW($1,$2,$3,$4,$5,$6,$7,$8)::kb.article_type
	)`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = tx.ExecContext(ctx, procedure,
		uuid.New(),
		article.Url,
		article.Title,
		article.Author,
		article.Published,
		article.Provider,
		article.Domain,
		article.Content,
	)

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	} else {
		fmt.Println(fmt.Sprintf("request %s committed", requestId))
	}

	return nil
}
