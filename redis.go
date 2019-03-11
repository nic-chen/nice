package nice

import (
	redislib "github.com/gomodule/redigo/redis"
	"github.com/mitchellh/mapstructure"
	"log"
	"time"
)

// redis is RDS struct
type Redis struct {
	host     string
	password string
	database int
	idle     int
	active   int
	pool     *redislib.Pool
}

type RedisCnf struct {
	//map类型
	Master RedisNode `yaml:"master"`
	Slave  RedisNode `yaml:"slave"`
}

type RedisNode struct {
	Host     string `yaml: "host"`
	Password string `yaml: "password"`
	Index    int    `yaml: "index"`
	MaxOpen  int    `yaml: "maxopen"` //maxOpenConn
	MaxIdle  int    `yaml: "maxidle"` //maxIdleConn
}

func NewRedis(config interface{}) *Redis {

	conf := MysqlConf{}
	err = mapstructure.Decode(config.(map[interface{}]interface{}), &conf)

	r := &Redis{
		host:     conf.Host,
		password: conf.Password,
		database: conf.Index,
		idle:     conf.MaxOpen,
		active:   conf.MaxIdle,
	}
	r.Open()
	if _, err := r.Do("PING"); err != nil {
		log.Panicln("Init redis pool failed.", err.Error())
	}
	return r
}

func (r *Redis) Open() {
	r.pool = &redislib.Pool{
		MaxActive:   r.active, // max number of connections
		MaxIdle:     r.idle,
		IdleTimeout: 120 * time.Second,
		Dial: func() (redislib.Conn, error) {
			c, err := redislib.Dial("tcp", r.host)
			if err != nil {
				return nil, err
			}
			if len(r.password) > 0 {
				if _, err := c.Do("AUTH", r.password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if _, err := c.Do("select", r.database); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redislib.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// Close pool
func (r *Redis) Close() error {
	err := r.pool.Close()
	return err
}

// Do commands
func (r *Redis) Do(command string, args ...interface{}) (interface{}, error) {
	conn := r.pool.Get()
	defer conn.Close()
	return conn.Do(command, args...)
}
