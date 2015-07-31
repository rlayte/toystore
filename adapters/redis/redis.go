package redis

import radix "github.com/fzzy/radix/redis"

type RedisStore struct {
	client *radix.Client
}

func (r RedisStore) Get(key string) (string, bool) {
	response := r.client.Cmd("GET", key)
	if response.Type == radix.NilReply {
		return "", false // Empty string is just a holder?
	}
	value, err := response.Str()
	if err != nil {
		panic(err)
	}
	return value, true
}

func (r RedisStore) Put(key string, value string) bool {
	err := r.client.Cmd("SET", key, value).Err
	if err != nil {
		panic(err)
	}

	return err == nil
}

func New(url string) *RedisStore {
	client, err := radix.Dial("tcp", url)
	if err != nil {
		panic(err)
	}
	return &RedisStore{client}
}

func (r RedisStore) Keys() []string {
	values := r.client.Cmd("KEYS")
	elems := values.Elems

	output := make([]string, len(elems))
	i := 0
	for _, e := range elems {
		key, err := e.Str()
		if err != nil {
			panic(err)
		}
		output[i] = key
		i++
	}
	return output
}
