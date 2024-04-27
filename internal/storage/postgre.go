package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (p *Postgre) AddURL(ctx context.Context, data URLData) error {
	p.Log.Debug("Добавляем URL в базу данных", zap.String("url", data.URL))
	id := ctx.Value(UserID)
	p.Log.Debug("Id из контекста", zap.Any("id", id))

	c, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()

	_, err := p.db.Query(c, "AddURL", data.ID, data.URL, id)
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
	id := ctx.Value(UserID)
	p.Log.Debug("Id из контекста", zap.Any("id", id))

	for _, v := range *b {
		// если данные не валидны, пропускаем текущую итерацию
		if v.Err != nil {
			continue
		}

		p.Log.Debug("Пытаемся добавить URL в БД", zap.String("url", v.URL), zap.String("id", v.ID))

		_, err := tx.Exec(ctx, "AddURL", v.ID, v.URL, id)

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

	c, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()

	insertType := fmt.Sprintf("SELECT id, url FROM urls WHERE %v = $1", t)
	err := p.db.QueryRow(c, insertType, v).Scan(&i, &u)
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
	c, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()

	// получаем данные для составления запроса
	val, queryInsert := p.getQueryInsert(ctx, urls)

	// Проверка существования урлов в базе данных
	query := "SELECT url, id FROM urls WHERE url IN (" + queryInsert + ")"
	rows, err := p.db.Query(c, query, val...)
	if err != nil {
		p.Log.Error("Error querying database:", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		var id string
		err := rows.Scan(&url, &id)
		if err != nil {
			p.Log.Error("Error scanning row:", zap.Error(err))
			return nil, err
		}
		for i, v := range *urls {
			if v.URL == url {
				(*urls)[i].Err = entity.ErrURLExist
				(*urls)[i].ID = id
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

func (p *Postgre) RemoveURL(_ context.Context, _ URLData) error {
	// TODO implement me
	panic("implement me4")
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
	rows, err := p.db.Query(c, "SELECT url, id FROM urls WHERE userID = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var url string
		var shortID string
		err := rows.Scan(&url, &shortID)
		if err != nil {
			p.Log.Error("Error scanning row:", zap.Error(err))
			return nil, err
		}
		urls = append(urls, URLData{
			URL: url,
			ID:  shortID,
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
	_, err = db.Prepare(ctx, "AddURL", `INSERT INTO urls (id, url, userID) VALUES ($1, $2, $3)`)
	if err != nil {
		panic(err)
	}
	_, err = db.Prepare(ctx, "GetNewUser", `INSERT INTO users DEFAULT VALUES RETURNING id`)
	if err != nil {
		panic(err)
	}
}
