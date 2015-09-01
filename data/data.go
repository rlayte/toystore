package data

import "time"

type Data struct {
	Key       string
	Value     interface{}
	Timestamp time.Time
}

func New(key string, value interface{}) *Data {
	return &Data{key, value, time.Now()}
}
