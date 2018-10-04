package mgo

import "sync"

const (
	TIME_FORMAT = "2006-01-02 15:04:05.000"
	SIZE_1K     = 1024
	SIZE_1M     = 1024 * 1024
)

var (
	once      = sync.Once{}
	UuidCache = make(map[int64]string)
	UuidMutex = sync.RWMutex{}
	Glogger   = (*Logger)(nil)
	Glimiter  = (*RateLimiter)(nil)
)
