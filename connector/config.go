package connector

type MysqlConfig struct {
	Path            string `json:"path" yaml:"path" mapstructure:"path"`                   // 服务器地址:端口
	WritePath       string `json:"write_path" yaml:"write_path" mapstructure:"write_path"` // 服务器地址:端口
	ReadPath        string `json:"read_path" yaml:"read_path" mapstructure:"read_path"`
	Config          string `json:"config" yaml:"config" mapstructure:"config"`                                  // 高级配置
	Dbname          string `json:"db_name" yaml:"db_name" mapstructure:"dbname"`                                // 数据库名
	Username        string `json:"username" yaml:"username" mapstructure:"username"`                            // 数据库用户名
	Password        string `json:"password" yaml:"password" mapstructure:"password"`                            // 数据库密码
	MaxIdleConns    int    `json:"max_idle_conns" yaml:"max_idle_conns" mapstructure:"max_idle_conns"`          // 空闲中的最大连接数
	MaxOpenConns    int    `json:"max_open_conns" yaml:"max_open_conns" mapstructure:"max_open_conns"`          // 打开到数据库的最大连接数
	ConnMaxLifetime int    `json:"conn_max_lifetime" yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"` // 空闲链接生命周期，单位秒钟
	DisableTrace    bool   `json:"disable_trace" yaml:"disable_trace" mapstructure:"disable_trace"`             // 是否会禁用 Trace
	DisableLog      bool   `json:"disable_log" yaml:"disable_log" mapstructure:"disable_log"`
}

type MongoConfig struct {
	Database     string `json:"database" yaml:"database" mapstructure:"database"`
	Address      string `json:"address" yaml:"address" mapstructure:"address"`
	Username     string `json:"username" yaml:"username" mapstructure:"username"`
	Password     string `json:"password" yaml:"password" mapstructure:"password"`
	EnableTLS    bool   `json:"enable_tls" yaml:"enable_tls" mapstructure:"enable_tls"` // 是否开启 tls
	Cfg          string `json:"cfg" yaml:"cfg" mapstructure:"cfg"`
	DisableTrace bool   `json:"disable_trace" yaml:"disable_trace" mapstructure:"disable_trace"` // 是否会禁用 Trace
	DisableLog   bool   `json:"disable_log" yaml:"disable_log" mapstructure:"disable_log"`
}

type RedisConfig struct {
	DB           int    `json:"db" yaml:"db" mapstructure:"db"`                                  // redis的哪个数据库
	Addr         string `json:"addr" yaml:"addr" mapstructure:"addr"`                            // 服务器地址:端口
	Username     string `json:"username" yaml:"username" mapstructure:"username"`                // 用户名
	Password     string `json:"password" yaml:"password" mapstructure:"password"`                // 密码
	EnableTLS    bool   `json:"enable_tls" yaml:"enable_tls" mapstructure:"enable_tls"`          // 是否开启 tls
	IsCluster    bool   `json:"is_cluster" yaml:"is_cluster" mapstructure:"is_cluster"`          // 是否是集群模式
	MasterOnly   bool   `json:"master_only" yaml:"master_only" mapstructure:"master_only"`       // 是否只读主库，仅在集群模式下生效
	PoolSize     int    `json:"pool_size" yaml:"pool_size" mapstructure:"pool_size"`             // 连接池大小
	DisableTrace bool   `json:"disable_trace" yaml:"disable_trace" mapstructure:"disable_trace"` // 是否会禁用 Trace
	EnableLog    bool   `json:"enable_log" yaml:"enable_log" mapstructure:"enable_log"`
}
