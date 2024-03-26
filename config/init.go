package config

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"strings"
)

const (
	runModeKey = "RUN_MODE"

	runModeProd = "prod"
	runModeTest = "test"
)

var Conf Config

func init() {
	Conf = loadConfig()
}

func getRunMode() string {
	switch modeInEnv := os.Getenv(runModeKey); modeInEnv {
	case runModeProd, runModeTest:
		return modeInEnv
	case "":
		return runModeTest
	default:
		log.Fatalf(
			"从环境变量读取%s值为%s。目前支持的值为:'%s' '%s'，不填默认为%s\n",
			runModeKey,
			modeInEnv,
			runModeProd,
			runModeTest,
			runModeTest,
		)
	}
	return runModeTest
}

func getConfigFilename(runMode string) string {
	_, thisFileName, _, _ := runtime.Caller(1)
	aPath := strings.Split(thisFileName, "/")
	dir := strings.Join(aPath[0:len(aPath)-1], "/")

	var jsonFilename string
	if runMode == runModeProd {
		jsonFilename = "config_prod.json"
	} else {
		jsonFilename = "config_dev.json"
	}

	return dir + "/" + jsonFilename
}

func loadConfig() Config {
	runMode := getRunMode()
	configPath := getConfigFilename(runMode)

	content, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("读取配置文件%s失败：%v\n", configPath, err)
	}

	var conf Config
	if err := json.Unmarshal(content, &conf); err != nil {
		log.Fatalf("json解析配置内容失败：%s\n", err)
	}

	logConfig(conf)

	return conf
}

func logConfig(conf Config) {
	log.Printf("Debug为：%t\n", conf.Debug)
	log.Printf("HTTP服务启动端口为：%d\n", conf.Server.Port)
	log.Printf("数据库组件为：%s\n", conf.Server.DbType)
	log.Printf("缓存组件为：%s\n", conf.Server.CacheType)
}
