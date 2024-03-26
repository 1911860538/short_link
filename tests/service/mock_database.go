package service

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/1911860538/short_link/app/component"
	"github.com/1911860538/short_link/config"
)

// 模拟数据库操作
type mockDatabase struct {
	mockLifespan

	mu    sync.Mutex
	maxId int
	db    map[string]*component.Link
}

func (t *mockDatabase) Create(ctx context.Context, link *component.Link) (id string, codeExisted bool, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 模拟code唯一索引已存在
	for _, dbLink := range t.db {
		if link.Code == dbLink.Code {
			return "", true, nil
		}
	}

	t.maxId++

	newId := strconv.Itoa(t.maxId)
	t.db[newId] = link

	return newId, false, nil
}

func (t *mockDatabase) Get(ctx context.Context, params map[string]any) (*component.Link, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, dbLink := range t.db {
		val := reflect.ValueOf(dbLink).Elem() // 获取s指向的结构体的反射值
		typ := val.Type()
		matched := true
		for k, v := range params {
			var field reflect.StructField
			fieldFound := false
			for i := 0; i < typ.NumField(); i++ {
				f := typ.Field(i)
				if tag := f.Tag.Get("bson"); tag == k {
					field = f
					fieldFound = true
					break
				}
			}
			if !fieldFound {
				return nil, fmt.Errorf("字段%s不存在", k)
			}

			fieldValue := val.FieldByName(field.Name)
			if !fieldValue.IsValid() || !fieldValue.Type().AssignableTo(reflect.TypeOf(v)) {
				return nil, fmt.Errorf("字段%s类型与数据库不匹配", k)
			}

			convertedValue := reflect.ValueOf(v).Convert(fieldValue.Type())
			if fieldValue.Interface() != convertedValue.Interface() {
				matched = false
				break
			}
		}
		if matched {
			return dbLink, nil
		}
	}
	return nil, nil
}

func (t *mockDatabase) latestLink() *component.Link {
	return t.db[strconv.Itoa(t.maxId)]
}

var _ component.DatabaseItf = (*mockDatabase)(nil)

// 构造测试数据
func getTestDb() (*mockDatabase, error) {

	keepTime := time.Duration(config.Conf.Core.ExpiredKeepDays*24) * time.Hour

	// 已过期数据
	link1Deadline := time.Now().UTC().Add(-time.Duration(2) * time.Hour)
	link1 := &component.Link{
		Id:        "1",
		UserId:    "user_id_1",
		Code:      "BIkBDO", // 通过code, err := service.GenCode(link1.UserId, link1.LongUrl, "")，获得
		Salt:      "",
		LongUrl:   "https://www.longurl1.com",
		Deadline:  link1Deadline,
		TtlTime:   link1Deadline.Add(keepTime),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 无过期时间数据
	link2 := &component.Link{
		Id:        "2",
		UserId:    "user_id_1",
		Code:      "2boMgt",
		Salt:      "",
		LongUrl:   "https://www.longurl2.com",
		Deadline:  time.Time{},
		TtlTime:   time.Time{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 有过期时间，但目前未过期数据
	link3Deadline := time.Now().UTC().Add(time.Duration(2) * time.Hour)
	link3 := &component.Link{
		Id:        "3",
		UserId:    "user_id_1",
		Code:      "fztcmW",
		Salt:      "",
		LongUrl:   "https://www.longurl3.com",
		Deadline:  link3Deadline,
		TtlTime:   link3Deadline.Add(keepTime),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 有过期时间，但目前未过期数据
	link4Deadline := time.Now().UTC().Add(time.Duration(30) * time.Hour)
	link4 := &component.Link{
		Id:        "4",
		UserId:    "user_id_1",
		Code:      "rCeB3h",
		Salt:      "",
		LongUrl:   "https://www.longurl4.com",
		Deadline:  link4Deadline,
		TtlTime:   link4Deadline.Add(keepTime),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	db := &mockDatabase{
		mockLifespan: mockLifespan{},
		db: map[string]*component.Link{
			"1": link1,
			"2": link2,
			"3": link3,
			"4": link4,
		},
	}
	db.maxId = len(db.db) + 1

	return db, nil
}
