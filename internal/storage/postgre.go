package storage

import (
	"context"
	"fmt"
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

func (p *Postgre) Ping() error {
	fmt.Println("Вызов")
	p.Log.Debug("Проверяем статус соединения с БД")
	//ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	//defer cancel()
	if err := p.db.Ping(context.Background()); err != nil {
		p.Log.Debug("Ошибка соединения с БД")
		fmt.Println(err)
		return err
	}
	p.Log.Debug("Есть соединение с БД")
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
