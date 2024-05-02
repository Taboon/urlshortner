package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/entity"
	pgx "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // postgres driver
	"github.com/pressly/goose"
	"go.uber.org/zap"
)

type Postgre struct {
	db  *pgx.Conn
	Log *zap.Logger
}

var _ Repository = (*Postgre)(nil)

func NewPostgreBase(db *pgx.Conn, log *zap.Logger) *Postgre {
	return &Postgre{
		db:  db,
		Log: log,
	}
}

func Migrations(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, "./")
}

func (p *Postgre) Ping(ctx context.Context) error {
	p.Log.Debug("Проверяем статус соединения с БД")
	c, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if err := p.db.Ping(c); err != nil {
		p.Log.Error("Ошибка соединения с БД")
		return err
	}
	p.Log.Info("Есть соединение с БД")
	return nil
}

func (p *Postgre) AddURL(ctx context.Context, urlData URLData) error {
	p.Log.Debug("Добавляем URL в базу данных", zap.String("url", urlData.URL))
	id := ctx.Value(UserID)
	p.Log.Debug("ID из контекста", zap.Any("id", id))
	deleted := false

	c, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()

	rows, err := p.db.Query(c, "AddURL", urlData.ID, urlData.URL, deleted, id)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (p *Postgre) WriteBatchURL(ctx context.Context, b *ReqBatchURLs) (*ReqBatchURLs, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	id := ctx.Value(UserID)
	p.Log.Debug("ID из контекста", zap.Any("id", id))
	deleted := false

	for _, v := range *b {
		// если данные не валидны, пропускаем текущую итерацию
		if v.Err != nil {
			continue
		}

		p.Log.Debug("Пытаемся добавить URL в БД", zap.String("url", v.URL), zap.String("id", v.ID))

		_, err := tx.Exec(ctx, "AddURL", v.ID, v.URL, deleted, id)

		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				return nil, err
			}
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return b, nil
}

// (doneCh chan struct{}, inputCh chan int) chan int
func (p *Postgre) CheckID(ctx context.Context, id string) (URLData, bool, error) {
	return p.check(ctx, "id", id)
}

func (p *Postgre) CheckURL(ctx context.Context, url string) (URLData, bool, error) {
	return p.check(ctx, "url", url)
}

func (p *Postgre) check(ctx context.Context, t string, v string) (URLData, bool, error) {
	var returnID string
	var returnURL string
	var deleted bool
	userID := ctx.Value(UserID)
	p.Log.Debug("Проверяем в базе", zap.Any("user", userID), zap.String("parametr", v))

	c := context.WithoutCancel(ctx)

	var err error
	if userID == 0 {
		insertType := fmt.Sprintf("SELECT id, url, deleted FROM urls WHERE %v = $1", t)
		err = p.db.QueryRow(c, insertType, v).Scan(&returnID, &returnURL, &deleted)
	} else {
		insertType := fmt.Sprintf("SELECT id, url, deleted FROM urls WHERE %v = $1 AND userid = $2", t)
		err = p.db.QueryRow(c, insertType, v, userID).Scan(&returnID, &returnURL, &deleted)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			p.Log.Debug("Не нашли запись в базе данных")
			return URLData{}, false, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			p.Log.Error("Другая ошибка запроса", zap.Error(err))
			return URLData{}, false, err
		}
	}
	p.Log.Debug("Возвращаем URLData", zap.String("url", returnURL), zap.String("id", returnID))
	return URLData{URL: returnURL, ID: returnID, Deleted: deleted}, true, nil
}

func (p *Postgre) CheckBatchURL(ctx context.Context, urls *ReqBatchURLs) (*ReqBatchURLs, error) { //nolint: funlen
	c, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()

	// получаем данные для составления запроса
	val, queryInsert := p.getQueryInsert(ctx, urls)

	// Проверка существования урлов в базе данных
	query := "SELECT url, id, deleted FROM urls WHERE url IN (" + queryInsert + ")"
	rows, err := p.db.Query(c, query, val...)
	if err != nil {
		p.Log.Error("Error querying database:", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		var id string
		var deleted bool
		err := rows.Scan(&url, &id, &deleted)
		if err != nil {
			p.Log.Error("Error scanning row:", zap.Error(err))
			return nil, err
		}
		for i, v := range *urls {
			if v.URL == url {
				(*urls)[i].Err = entity.ErrURLExist
				(*urls)[i].ID = id
				(*urls)[i].Deleted = deleted
			}
		}
	}
	return urls, nil
}

func (p *Postgre) getQueryInsert(_ context.Context, urls *ReqBatchURLs) ([]interface{}, string) {
	val := make([]interface{}, 0, len(*urls))
	var queryInsert string
	i := 0
	for _, v := range *urls {
		if v.Err != nil {
			continue
		}
		val = append(val, v.URL)
		queryInsert += fmt.Sprintf("$%v", i+1)
		queryInsert += ","
		i++
	}

	queryInsert = strings.TrimSuffix(queryInsert, ",")
	return val, queryInsert
}

func (p *Postgre) RemoveURL(ctx context.Context, data []URLData) error {
	remove := true
	c := context.WithoutCancel(ctx)
	tx, err := p.db.Begin(context.Background())
	if err != nil {
		log.Fatalf("Unable to begin transaction: %v", err)
	}

	defer func() {
		if err != nil {
			err := tx.Rollback(context.Background())
			if err != nil {
				p.Log.Error("Rollback transaction failed", zap.Error(err))
			}
			p.Log.Error("Transaction failed", zap.Error(err))
		} else {
			err := tx.Commit(context.Background())
			if err != nil {
				p.Log.Error("Commit transaction failed", zap.Error(err))
			}
		}
	}()

	for _, url := range data {
		_, err := tx.Exec(c, "UPDATE urls SET deleted = $1 WHERE id = $2", remove, url.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Postgre) GetNewUser(ctx context.Context) (int, error) {
	c, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()

	p.Log.Debug("Добавляем пользователя в базу и получаем ID")
	var id int
	err := p.db.QueryRow(c, "GetNewUser").Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (p *Postgre) GetURLsByUser(ctx context.Context, id int) (UserURLs, error) {
	c, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	p.Log.Debug("Получаем все URL пользователя", zap.Int("id", id))

	urls := UserURLs{}
	rows, err := p.db.Query(c, "SELECT url, id, deleted FROM urls WHERE userID = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var url string
		var shortID string
		var deleted bool
		err := rows.Scan(&url, &shortID)
		if err != nil {
			p.Log.Error("Error scanning row:", zap.Error(err))
			return nil, err
		}
		urls = append(urls, URLData{
			URL:     url,
			ID:      shortID,
			Deleted: deleted,
		})
	}
	return urls, nil
}

func SetPostgres(ctx context.Context, conf *config.Config, l *zap.Logger) (*pgx.Conn, Repository) {
	db, err := pgx.Connect(ctx, conf.DataBase)

	if err != nil {
		fprintf, err := fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		if err != nil {
			return nil, nil
		}
		panic(fprintf)
	}

	stor := NewPostgreBase(db, l)
	err = stor.Ping(ctx)
	if err != nil {
		fprintf, err := fmt.Fprintf(os.Stderr, "Can't connect to database: %v\n", err)
		if err != nil {
			return nil, nil
		}
		panic(fprintf)
	}

	err = Migrations(conf.DataBase)
	if err != nil {
		fprintf, err := fmt.Fprintf(os.Stderr, "Can't created table: %v\n", err)
		if err != nil {
			return nil, nil
		}
		panic(fprintf)
	}

	setPrepare(db)

	return db, stor
}

func setPrepare(db *pgx.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	_, err := db.Prepare(ctx, "GetURLSByUser", `SELECT url, id FROM urls WHERE userID = $1`)
	if err != nil {
		panic(err)
	}
	_, err = db.Prepare(ctx, "AddURL", `INSERT INTO urls (id, url, deleted, userID) VALUES ($1, $2, $3, $4)`)
	if err != nil {
		panic(err)
	}
	_, err = db.Prepare(ctx, "GetNewUser", `INSERT INTO users DEFAULT VALUES RETURNING id`)
	if err != nil {
		panic(err)
	}
}
