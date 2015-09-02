package data

import "time"

type Data struct {
	Key       string
	Value     interface{}
	Timestamp time.Time
}

func (d *Data) IsLater(other *Data) bool {
	return d.Timestamp.After(other.Timestamp)
}

func New(key string, value interface{}) *Data {
	return &Data{key, value, time.Now()}
}
