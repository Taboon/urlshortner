package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/entity"
	"github.com/pressly/goose"
	"os"
	"strings"
	"time"

	pgx "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // postgres driver
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

	return goose.Up(db, "internal/db/migrations")
}

func (p *Postgre) Ping() error {
	p.Log.Debug("Проверяем статус соединения с БД")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := p.db.Ping(ctx); err != nil {
		p.Log.Error("Ошибка соединения с БД")
		return err
	}
	p.Log.Info("Есть соединение с БД")
	return nil
}

func (p *Postgre) AddURL(ctx context.Context, data URLData) error {
	p.Log.Debug("Добавляем URL в базу данных", zap.String("url", data.URL))
	_, err := p.db.Exec(ctx, "INSERT INTO urls (id, url) VALUES ($1, $2)", data.ID, data.URL)
	if err != nil {
		return err
	}
	return nil
}

func (p *Postgre) WriteBatchURL(ctx context.Context, b *ReqBatchURLs) (*ReqBatchURLs, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	for _, v := range *b {
		// если данные не валидны, пропускаем текущую итерацию
		if v.Err != nil {
			continue
		}

		p.Log.Debug("Пытаемся добавить URL в БД", zap.String("url", v.URL), zap.String("id", v.ID))

		_, err := tx.Exec(ctx, "INSERT INTO urls (id, url) VALUES($1, $2)", v.ID, v.URL)
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

func (p *Postgre) CheckID(ctx context.Context, id string) (URLData, bool, error) {
	return p.check(ctx, "id", id)
}

func (p *Postgre) CheckURL(ctx context.Context, url string) (URLData, bool, error) {
	return p.check(ctx, "url", url)
}

func (p *Postgre) check(ctx context.Context, t string, v string) (URLData, bool, error) {
	var i string
	var u string

	insertType := fmt.Sprintf("SELECT id, url FROM urls WHERE %v = $1", t)
	err := p.db.QueryRow(ctx, insertType, v).Scan(&i, &u)
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
	p.Log.Debug("Возвращаем URLData", zap.String("url", u), zap.String("id", i))
	return URLData{URL: u, ID: i}, true, nil
}

func (p *Postgre) CheckBatchURL(ctx context.Context, urls *ReqBatchURLs) (*ReqBatchURLs, error) {
	// получаем данные для составления запроса
	val, queryInsert := p.getQueryInsert(urls)

	// Проверка существования урлов в базе данных
	query := "SELECT url FROM urls WHERE url IN (" + queryInsert + ")"
	rows, err := p.db.Query(ctx, query, val...)
	if err != nil {
		p.Log.Error("Error querying database:", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		err := rows.Scan(&url)
		if err != nil {
			p.Log.Error("Error scanning row:", zap.Error(err))
			return nil, err
		}
		for i, v := range *urls {
			if v.URL == url {
				(*urls)[i].Err = entity.ErrURLExist
			}
		}
	}
	return urls, nil
}

func (p *Postgre) getQueryInsert(urls *ReqBatchURLs) ([]interface{}, string) {
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

func (p *Postgre) RemoveURL(_ context.Context, _ URLData) error {
	// TODO implement me
	panic("implement me4")
}

func SetPostgres(conf *config.Config, l *zap.Logger) (*pgx.Conn, Repository) {
	db, err := pgx.Connect(context.Background(), conf.DataBase)

	if err != nil {
		fprintf, err := fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		if err != nil {
			return nil, nil
		}
		panic(fprintf)
	}

	stor := NewPostgreBase(db, l)
	err = stor.Ping()
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

	l.Info("Использем Postge")
	return db, stor
}
