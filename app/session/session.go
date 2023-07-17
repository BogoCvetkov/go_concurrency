package session

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/bogo/go-concurrency/data"
	"github.com/gomodule/redigo/redis"
)

var sessionManager *scs.SessionManager

func InitializeSessionManager() *scs.SessionManager {
	gob.Register(data.User{})

	// Establish connection pool to Redis.
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}

	// Initialize a new session manager and configure it to use redisstore as the session store.
	sessionManager = scs.New()
	sessionManager.Store = redisstore.New(pool)
	sessionManager.Lifetime = 3 * time.Hour
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = true

	return sessionManager
}
