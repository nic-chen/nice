package nice

import (
	"log"
	"time"
	redislib "github.com/gomodule/redigo/redis"
)

// redis is RDS struct
type Redis struct {
	host string
	password string
	database int
	idle int
	active int
	pool *redislib.Pool
}

func NewRedis(host, password string, database, MaxActive, maxIdleConns int) *Redis {
	r := &Redis{
		host: host,
		password: password,
		database: database,
		idle: maxIdleConns,
		active: MaxActive,
	};
	r.Open()
	if _, err := r.Do("PING"); err != nil {
		log.Panicln("Init redis pool failed.", err.Error())
	}
	return r
}

func (r *Redis) Open(){
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
	};
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
