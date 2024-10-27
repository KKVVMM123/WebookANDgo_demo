package domain

import "time"

// User 领域对象 是DDD中的entity
type User struct {
	Id       int64
	Email    string
	Password string
	Ctime    time.Time
}
