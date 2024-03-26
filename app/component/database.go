package component

import (
	"context"
	"log"

	"github.com/1911860538/short_link/config"
)

type DatabaseItf interface {
	Lifespan

	Create(ctx context.Context, link *Link) (id string, codeExisted bool, err error)
	Get(ctx context.Context, params map[string]any) (*Link, error)
}

var Database DatabaseItf

func init() {
	switch databaseType := config.Conf.Server.DbType; databaseType {
	case "mongodb":
		Database = DefaultMongoDB
	default:
		log.Fatalf("不支持的数据库组件：%s\n", databaseType)
	}
}
