package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"time"
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

func CheckDB(db *pgx.Conn) error {
	_, err := db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS urls (id TEXT PRIMARY KEY, url TEXT)")
	if err != nil {
		return err
	}
	return nil
}

func (p *Postgre) Ping() error {
	p.Log.Debug("Проверяем статус соединения с БД")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := p.db.Ping(ctx); err != nil {
		p.Log.Error("Ошибка соединения с БД")
		fmt.Println(err)
		return err
	}
	p.Log.Info("Есть соединение с БД")
	return nil
}

func (p *Postgre) AddURL(data URLData) error {
	p.Log.Debug("Добавляем URL в базу данных", zap.String("url", data.URL))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := p.db.Exec(ctx, "INSERT INTO urls (id, url) VALUES ($1, $2)", data.ID, data.URL)
	if err != nil {
		p.Log.Error("Ошибка при добавлении нового пользователя", zap.Error(err))
		return err
	}
	return nil
}

func (p *Postgre) CheckID(id string) (URLData, bool, error) {
	var i string
	var u string
	p.Log.Debug("Проверяем ID в базе данных", zap.String("id", id))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := p.db.QueryRow(ctx, "SELECT id, url FROM urls WHERE id = $1", id).Scan(&i, &u)
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

func (p *Postgre) CheckURL(url string) (URLData, bool, error) {
	var i string
	var u string
	p.Log.Debug("Проверяем URL в базе данных", zap.String("url", url))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := p.db.QueryRow(ctx, "SELECT id, url FROM urls WHERE url = $1", url).Scan(&i, &u)
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

func (p *Postgre) RemoveURL(data URLData) error {
	//TODO implement me
	panic("implement me4")
}
