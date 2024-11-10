package cache

import (
	"sync"
	"time"
)

type ICache interface {
	Get(key string) *Client
	Set(key string, client *Client)

	GetClient(footprint string) *Props
	SetClient(footprint string, client *Props)
}

type Props struct {
	// rate limiting
	RateLimitRemain uint8     `json:"rateLimitRemain"`
	LastReq         time.Time `json:"lastReq"`

	// Inner properties
	LastUpdated time.Time `json:"lastUpdated"`
}

type Client struct {
	IAID string `json:"IAID"`

	// Client properties
	IsBanned bool   `json:"isBanned"`
	Login    string `json:"login"`

	Props Props `json:"props"`
}

type Cache struct {
	UserStorage   map[string]*Client
	ClientStorage map[string]*Props
	URWm          *sync.RWMutex
	CRWm          *sync.RWMutex
	TempUsers     []string
}

func NewCache() *Cache {
	return &Cache{
		UserStorage:   make(map[string]*Client),
		ClientStorage: make(map[string]*Props),
		URWm:          new(sync.RWMutex),
		CRWm:          new(sync.RWMutex),
		TempUsers:     make([]string, 0),
	}
}

func (c *Cache) Get(key string) *Client {
	c.URWm.RLock()
	defer c.URWm.RUnlock()
	if client, ok := c.UserStorage[key]; ok {
		return client
	}

	return nil
}

func (c *Cache) Set(key string, client *Client) {
	c.URWm.Lock()
	defer c.URWm.Unlock()

	c.UserStorage[key] = client
}

func (c *Cache) GetClient(footprint string) *Props {
	c.CRWm.RLock()

	defer c.CRWm.RUnlock()

	if _, ok := c.ClientStorage[footprint]; !ok {
		return nil
	}

	return c.ClientStorage[footprint]
}

func (c *Cache) SetClient(footprint string, props *Props) {
	c.CRWm.Lock()
	defer c.CRWm.Unlock()

	c.ClientStorage[footprint] = props
}
