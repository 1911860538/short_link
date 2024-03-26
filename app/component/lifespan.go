package component

import (
	"log"
	"sync"
)

// Lifespan 约束组件（缓存、数据库等），必须实现初始化和停用的方法
type Lifespan interface {
	Startup() error
	Shutdown() error
}

// components 所有组件的容器
type components struct {
	items []Lifespan
}

func (c *components) register(ls Lifespan) {
	if c.items == nil {
		c.items = make([]Lifespan, 0)
	}
	c.items = append(c.items, ls)
}

var defaultCs = &components{}

func init() {
	defaultCs.register(Cache)
	defaultCs.register(Database)
}

func (c *components) startup() error {
	for i := range c.items {
		if err := c.items[i].Startup(); err != nil {
			return err
		}
	}
	return nil
}

func (c *components) shutdown() {
	for i := range c.items {
		if err := c.items[i].Shutdown(); err != nil {
			log.Printf("关闭资源错误：%v\n", err)
		}
	}
}

var (
	onceStartUp  sync.Once
	onceShutdown sync.Once
)

// Startup 项目启动，初始化所有组件
func Startup() error {
	var err error
	onceStartUp.Do(func() {
		err = defaultCs.startup()
	})
	return err
}

// Shutdown 项目停止运行，对组件连接等关闭、资源清理
func Shutdown() {
	onceShutdown.Do(func() {
		defaultCs.shutdown()
	})
}
