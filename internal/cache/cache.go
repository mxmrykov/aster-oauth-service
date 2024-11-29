package cache

import (
	"encoding/base64"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type ICache interface {
	Get(key string) *Client
	Set(key string, client *Client)

	GenSignature()
	GetSignature() string

	GetClient(footprint string) *Props
	SetClient(footprint string, client *Props)

	MapAllCl() map[string]*Props
}

type Props struct {
	ASID string `json:"ASID"`

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

type signature struct {
	Signature string
	Exp       time.Time
}

type Cache struct {
	CurrentSignature *signature
	UserStorage      map[string]*Client
	ClientStorage    map[string]*Props
	URWm             *sync.RWMutex
	CRWm             *sync.RWMutex
	TempUsers        []string
}

func NewCache() *Cache {
	return &Cache{
		CurrentSignature: new(signature),
		UserStorage:      make(map[string]*Client),
		ClientStorage:    make(map[string]*Props),
		URWm:             new(sync.RWMutex),
		CRWm:             new(sync.RWMutex),
		TempUsers:        make([]string, 0),
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

func (c *Cache) GenSignature() {
	c.sign()
	<-time.After(time.Minute)
}

func (c *Cache) GetSignature() string {
	return c.CurrentSignature.Signature
}

func (c *Cache) sign() {
	pass, z := c.CurrentSignature.Exp.Before(time.Now()), c.CurrentSignature.Exp.IsZero()
	if pass && !z {
		return
	}

	ns := new(signature)
	ns.Exp = time.Now().Add(1 * time.Minute)
	s := base64.StdEncoding.EncodeToString(
		[]byte(
			uuid.New().String(),
		),
	)

	ns.Signature = strings.ToUpper(s)

	log.Info().Msgf("New signature: %s, exp: %v", ns.Signature, ns.Exp)
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

func (c *Cache) MapAllCl() map[string]*Props {
	c.CRWm.RLock()

	defer c.CRWm.RUnlock()

	return c.ClientStorage
}
