package storage

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"time"
)

type Postgre struct {
	db  *sql.DB
	Log *zap.Logger
}

var _ Repository = (*Postgre)(nil)

func NewPostgreBase(name string, login string, pass string, ip string, port string, log *zap.Logger) *Postgre {
	ps := fmt.Sprintf("host=%s:%s user=%s password=%s dbname=%s sslmode=disable", ip, port, login, pass, name)
	fmt.Println(ps)
	db, err := sql.Open("pgx", ps)
	if err != nil {
		panic(err)
	}
	return &Postgre{
		db:  db,
		Log: log,
	}
}

func (p *Postgre) Ping() error {
	p.Log.Info("Проверяем статус соединения с БД")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := p.db.PingContext(ctx); err != nil {
		p.Log.Info("Ошибка соединения с БД")
		return err
	}
	p.Log.Info("Есть соединение с БД")
	return nil
}

func (p *Postgre) AddURL(data URLData) error {
	//TODO implement me
	panic("implement me1")
}

func (p *Postgre) CheckID(id string) (URLData, bool, error) {
	//TODO implement me
	panic("implement me2")
}

func (p *Postgre) CheckURL(url string) (URLData, bool, error) {
	//TODO implement me
	panic("implement me3")
}

func (p *Postgre) RemoveURL(data URLData) error {
	//TODO implement me
	panic("implement me4")
}
