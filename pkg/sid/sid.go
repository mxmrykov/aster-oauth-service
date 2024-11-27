package sid

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Sid struct {
	Checksum string `json:"checksum"`

	Subscriber string `json:"subscriber"`

	SignDt  time.Time `json:"signDt"`
	Expires time.Time `json:"expires"`
}

func New(iaid string, dur time.Duration) string {
	n := time.Now()
	cs := md5.Sum(
		[]byte(fmt.Sprintf("%s_%d", iaid, n.Unix())),
	)

	s, _ := json.Marshal(Sid{
		Subscriber: iaid,
		Checksum:   hex.EncodeToString(cs[:]),
		SignDt:     n,
		Expires:    n.Add(dur),
	})

	return base64.StdEncoding.EncodeToString(s)
}

func Validate(sid string) (*Sid, error) {
	r, err := base64.StdEncoding.DecodeString(sid)

	if err != nil {
		return nil, fmt.Errorf("cannot decode sid: %s", err.Error())
	}

	m := new(Sid)

	if err = json.Unmarshal(r, &m); err != nil {
		return nil, fmt.Errorf("cannot unmarshal sid details: %s", err.Error())
	}

	cs := md5.Sum(
		[]byte(fmt.Sprintf("%s_%d", m.Subscriber, m.SignDt.Unix())),
	)

	switch {
	case m.Expires.Before(time.Now()):
		return nil, fmt.Errorf("sid is expired")
	case m.SignDt.After(time.Now()):
		return nil, fmt.Errorf("sid is not valid yet")
	case m.Checksum == "":
		return nil, fmt.Errorf("invalid checksum")
	case m.Checksum != hex.EncodeToString(cs[:]):
		return nil, fmt.Errorf("invalid checksum")
	}

	return m, nil
}
