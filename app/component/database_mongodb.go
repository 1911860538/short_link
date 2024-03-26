package component

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/1911860538/short_link/config"
)

type MongoDB struct {
	client *mongo.Client
}

var _ DatabaseItf = (*MongoDB)(nil)

var DefaultMongoDB = &MongoDB{}

func (m *MongoDB) Startup() error {
	connTimeout := time.Duration(config.Conf.MongoDB.ConnTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	var uri string
	if config.Conf.MongoDB.Username != "" {
		uri = fmt.Sprintf(
			"mongodb://%s:%s@%s:%d",
			config.Conf.MongoDB.Username,
			config.Conf.MongoDB.Password,
			config.Conf.MongoDB.Host,
			config.Conf.MongoDB.Port,
		)
	} else {
		uri = fmt.Sprintf(
			"mongodb://%s:%d",
			config.Conf.MongoDB.Host,
			config.Conf.MongoDB.Port,
		)
	}
	clientOptions := options.Client().ApplyURI(uri).SetConnectTimeout(connTimeout)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return err
	}
	log.Printf("成功连接mongodb")

	m.client = client

	return nil
}

func (m *MongoDB) Shutdown() error {
	if err := m.client.Disconnect(context.Background()); err != nil {
		return err
	}
	log.Printf("关闭mongodb连接")
	return nil
}

const (
	LinkCollectionName = "links"
)

// Link
/*
对于添加索引操作，官方go驱动不能在结构体tag赋值完成
需要在该collection创建了，并包含至少一个document，才能添加索引

code唯一索引： db.links.createIndex({"code": 1}, {"unique": true})
user_id普通索引： db.links.createIndex({"user_id": 1})
long_url普通索引： db.links.createIndex({"long_url": 1})
ttl_time ttl索引： db.links.createIndex({"ttl_time": 1}, {"expireAfterSeconds": 7200})
// ttl索引会增加数据库负载。如果不使用ttl索引，可以用定时脚本任务删除无用数据
*/
type Link struct {
	Id        string    `bson:"_id,omitempty"`
	UserId    string    `bson:"user_id"`
	Code      string    `bson:"code"`
	Salt      string    `bson:"salt"`
	LongUrl   string    `bson:"long_url"`
	Deadline  time.Time `bson:"deadline"`
	TtlTime   time.Time `bson:"ttl_time"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func (l *Link) Expired() bool {
	if l.Deadline.IsZero() {
		return false
	}

	return l.Deadline.Before(time.Now().UTC())
}

func (m *MongoDB) getLinkCollection() *mongo.Collection {
	return m.client.Database(config.Conf.MongoDB.DbName).Collection(LinkCollectionName)
}

func (m *MongoDB) Create(ctx context.Context, link *Link) (id string, existed bool, err error) {
	insertRes, err := m.getLinkCollection().InsertOne(ctx, link)

	// 判断错误类型是否为code已存在
	var writeErr mongo.WriteException
	if ok := errors.As(err, &writeErr); ok {
		if writeErr.HasErrorCode(11000) {
			return "", true, nil
		}
	}

	bytesId := insertRes.InsertedID.(primitive.ObjectID)
	return bytesId.Hex(), false, err
}

func (m *MongoDB) Get(ctx context.Context, params map[string]any) (*Link, error) {
	var link Link
	filter := make(bson.M, len(params))
	for k, v := range params {
		filter[k] = v
	}
	err := m.getLinkCollection().FindOne(ctx, filter).Decode(&link)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &link, nil
}
