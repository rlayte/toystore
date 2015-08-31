package toystore

import "time"

type Data struct {
	key       string
	value     interface{}
	timestamp time.Time
}

func NewData(key string, value interface{}) *Data {
	return &Data{key, value, time.Now()}
}
