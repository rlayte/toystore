// Package data is used to represent key/value pairs internally and transmit
// them over RPC between nodes.
package data

import (
	"fmt"
	"time"
)

// Data is used internally to store key/value pairs.
// A timestamp of when it was created is assigned to resolve data conflicts.
type Data struct {
	Key       string
	Value     interface{}
	Timestamp time.Time
}

// IsLater takes another Data item and compares their timestamps.
// If the other item is newer it will return false otherwise it will
// return true.
func (d *Data) IsLater(other *Data) bool {
	return d.Timestamp.After(other.Timestamp)
}

// String returns a string in the format key/value.
func (d *Data) String() string {
	return fmt.Sprintf("%s/%v", d.Key, d.Value)
}

// New creates a Data struct with the key/value provided and the current
// as its Timestamp.
func New(key string, value interface{}) *Data {
	return &Data{key, value, time.Now()}
}
