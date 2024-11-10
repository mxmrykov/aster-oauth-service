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

func New(iaid string) string {
	n := time.Now()
	cs := md5.Sum(
		[]byte(fmt.Sprintf("%s_%d", iaid, n.Unix())),
	)

	s, _ := json.Marshal(Sid{
		Subscriber: iaid,
		Checksum:   hex.EncodeToString(cs[:]),
		SignDt:     n,
		Expires:    n.Add(5 * time.Minute),
	})

	return base64.StdEncoding.EncodeToString(s)
}

func Validate(sid string) error {
	r, err := base64.StdEncoding.DecodeString(sid)

	if err != nil {
		return fmt.Errorf("cannot decode sid: %s", err.Error())
	}

	m := new(Sid)

	if err = json.Unmarshal(r, &m); err != nil {
		return fmt.Errorf("cannot unmarshal sid details: %s", err.Error())
	}

	cs := md5.Sum(
		[]byte(fmt.Sprintf("%s_%d", m.Subscriber, m.SignDt.Unix())),
	)

	switch {
	case m.Expires.Before(time.Now()):
		return fmt.Errorf("sid is expired")
	case m.SignDt.After(time.Now()):
		return fmt.Errorf("sid is not valid yet")
	case m.Checksum == "":
		return fmt.Errorf("invalid checksum")
	case m.Checksum != hex.EncodeToString(cs[:]):
		return fmt.Errorf("invalid checksum")
	}

	return nil
}
