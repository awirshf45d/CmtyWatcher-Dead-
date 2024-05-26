package config

// Database
type DBConfig struct {
	DBUsername string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}

var DBconf = DBConfig{
	DBUsername: "CmtyWatcher_User",
	DBPassword: "4587e5e94dd85bd013094eff36c02401e6fdf440ed5667f6a56b06c6ca03568b",
	DBHost:     "localhost",
	DBPort:     "3306",
	DBName:     "CmtyWatcher",
}

// Proxy
type ProxyComponents struct {
	HttpProxyURL string
}

var Proxies []string = []string{"http://127.0.0.1:2081"}
