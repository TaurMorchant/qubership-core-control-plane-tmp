package cache

type CacheClient struct {
	cache map[string]interface{}
}

func NewCacheClient() *CacheClient {
	return &CacheClient{
		cache: make(map[string]interface{}),
	}
}

func (client *CacheClient) Set(key string, value interface{}) {
	client.cache[key] = value
}

func (client *CacheClient) Delete(key string) {
	delete(client.cache, key)
}

func (client *CacheClient) Get(key string) (interface{}, bool) {
	value, ok := client.cache[key]
	return value, ok
}

func (client *CacheClient) GetAll() map[string]interface{} {
	return client.cache
}
