package netaddr

import (
	"errors"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port int64
}

func (na NetAddress) String() string {
	return na.Host + ":" + strconv.FormatInt(na.Port, 10)
}

func (na *NetAddress) UnmarshalText(text []byte) error {
	return na.Set(string(text))
}

func (na *NetAddress) Set(s string) error {
	if s == "" {
		return nil
	}

	s, _ = strings.CutPrefix(s, "http://")

	hp := strings.Split(s, ":")
	switch len(hp) {
	case 1:
		na.Host = hp[0]
	case 2:
		port, err := strconv.ParseInt(hp[1], 10, 0)
		if err != nil {
			return err
		}
		na.Host = hp[0]
		na.Port = port
	default:
		return errors.New("need address in a form host:port")
	}
	return nil
}
