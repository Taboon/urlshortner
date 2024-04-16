package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	//"github.com/Taboon/urlshortner/internal/domain/usecase"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
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

func (p *Postgre) AddBatchURL(urls map[string]ReqBatchJSON) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	for id, v := range urls {
		// все изменения записываются в транзакцию
		if v.Valid && !v.Exist {
			_, err := tx.Exec(ctx,
				"INSERT INTO urls (id, url)"+
					" VALUES($1, $2)", id, v.URL)
			if err != nil {
				// если ошибка, то откатываем изменения
				err := tx.Rollback(ctx)
				if err != nil {
					return err
				}
				return err
			}
		}
	}
	// завершаем транзакцию
	return tx.Commit(ctx)
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

func (p *Postgre) CheckBatchURL(urls *[]ReqBatchJSON) (*[]ReqBatchJSON, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Создание временной таблицы для урлов
	tx, err := p.db.Begin(ctx)
	if err != nil {
		fmt.Println("Error beginning transaction:", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback(context.Background())
		p.Log.Error("Ошибка RollBack", zap.Error(err))
	}()

	_, err = tx.Exec(context.Background(), "CREATE TEMP TABLE temp_urls (url TEXT)")
	if err != nil {
		fmt.Println("Error creating temporary table:", err)
		return nil, err
	}

	// Вставка урлов в временную таблицу
	var values []interface{}
	for _, v := range *urls {
		if v.Valid && !v.Exist {
			values = append(values, v.URL)
		}
	}

	var queryInsert string
	for i := 1; i <= len(values); i++ {
		queryInsert += fmt.Sprintf("($%v)", i)
		if i < len(values) {
			queryInsert += ","
		}
	}

	_, err = tx.Exec(context.Background(), "INSERT INTO temp_urls (url) VALUES "+queryInsert, values...)
	if err != nil {
		fmt.Println("Error inserting urls into temporary table:", err)
		return nil, err
	}

	// Проверка существования урлов в базе данных
	query := "SELECT url FROM temp_urls tu WHERE EXISTS (SELECT 1 FROM urls u WHERE u.url = tu.url)"
	rows, err := p.db.Query(context.Background(), query)
	if err != nil {
		fmt.Println("Error querying database:", err)
		return nil, err
	}
	defer rows.Close()

	existingUrls := make(map[string]bool)
	for rows.Next() {
		var url string
		err := rows.Scan(&url)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			return nil, err
		}
		existingUrls[url] = true
	}

	// Вывод результатов
	for i, url := range *urls {
		if existingUrls[url.URL] {
			p.Log.Info("URL уже есть в базе.", zap.String("url", url.URL))
			(*urls)[i].Exist = true
		} else {
			(*urls)[i].Exist = false
		}
	}
	return urls, nil
}

func (p *Postgre) RemoveURL(data URLData) error {
	//TODO implement me
	panic("implement me4")
}
