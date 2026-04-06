package pgsql

import (
	"bluebell/setting"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// 小写，不对外暴露
var db *sqlx.DB

func Init(cfg *setting.PostgresConfig) (err error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DbName,
	)
	db, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		zap.L().Error("connect to DB failed", zap.Error(err))
		return
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	return
}

// 小技巧
// 因为db小写，不对外暴露
// 可以封装一个Close
func Close() {
	_ = db.Close()
}
