package pgsql

import (
	"bluebell/models"
	"bluebell/setting"
	"testing"
	"time"
)

func init() {
	dbCfg := setting.PostgresConfig{
		Host:         "127.0.0.1",
		User:         "postgres",
		Password:     "123456",
		DbName:       "gamepulse",
		Port:         5432,
		MaxOpenConns: 10,
		MaxIdleConns: 10,
	}
	err := Init(&dbCfg)
	if err != nil {
		panic(err)
	}
}

func TestCreatePost(t *testing.T) {
	post := &models.Post{
		ID:          time.Now().UnixNano(),
		AuthorID:    123,
		CommunityID: 1,
		Title:       "test",
		Content:     "just a test",
	}
	err := CreatePost(post)
	if err != nil {
		t.Fatalf("CreatePost failed, err:%v\n", err)
	}
	t.Logf("CreatePost success")

}
