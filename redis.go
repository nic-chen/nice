package nice

import (
	"log"
	"time"
	redislib "github.com/gomodule/redigo/redis"
)

// redis is RDS struct
type redis struct {
	pool *redislib.Pool
}

func NewRedis(host, password string, database, maxOpenConns, maxIdleConns int) *redis {
	r := &redis{};
	r.Open(host, password, database, maxOpenConns, maxIdleConns)
	if _, err := r.Do("PING"); err != nil {
		log.Panicln("Init redis pool failed.", err.Error())
	}
	return r
}

func (p *redis) Open(server, password string, database, maxOpenConns, maxIdleConns int) {
	p.pool = &redislib.Pool{
		MaxActive:   maxOpenConns, // max number of connections
		MaxIdle:     maxIdleConns,
		IdleTimeout: 120 * time.Second,
		Dial: func() (redislib.Conn, error) {
			c, err := redislib.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if _, err := c.Do("select", database); err != nil {
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
func (p *redis) Close() error {
	err := p.pool.Close()
	return err
}

// Do commands
func (p *redis) Do(command string, args ...interface{}) (interface{}, error) {
	conn := p.pool.Get()
	defer conn.Close()
	return conn.Do(command, args...)
}
