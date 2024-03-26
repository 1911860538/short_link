package config

type Config struct {
	Debug bool `json:"debug"`

	Server  Server  `json:"server"`
	Jwt     Jwt     `json:"jwt"`
	Core    Core    `json:"core"`
	MongoDB MongoDB `json:"mongodb"`
	Redis   Redis   `json:"redis"`
}

type Server struct {
	Port               int    `json:"port"`
	IdleTimeoutSeconds int    `json:"idle_timeout_seconds"`
	DbType             string `json:"db_type"`
	CacheType          string `json:"cache_type"`
}

type Jwt struct {
	SecretKey string `json:"secret_key"`
	Algo      string `json:"algo"`
}

// Core 项目业务逻辑配置项，比如缓存key过期时间等
type Core struct {
	LongUrlConnTimeout int    `json:"long_url_conn_timeout"`
	CodeLen            int    `json:"code_len"`
	CodeTtl            int    `json:"code_ttl"`
	RedirectStatusCode int    `json:"redirect_status_code"`
	ExpiredKeepDays    int    `json:"expired_keep_days"`
	CacheNotFoundValue string `json:"cache_not_found_value"`
}

type MongoDB struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	DbName      string `json:"db_name"`
	ConnTimeout int    `json:"conn_timeout"`
}

type Redis struct {
	Password    string `json:"password"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Db          int    `json:"db"`
	ConnTimeout int    `json:"conn_timeout"`
}
