package bobo

import (
	"net/url"
	"strconv"
	"strings"
)

type Params url.Values

var truthy []string = []string{"true", "t", "yes", "y", "on", "1"}

func (p Params) Get(key string) string {
	if p == nil {
		return ""
	}

	vs, ok := p[key]
	if !ok || len(vs) == 0 {
		return ""
	}

	return vs[0]
}

func (p Params) Int64(key string) int64 {
	n, _ := strconv.ParseInt(p.Get(key), 10, 0)
	return n
}

func (p Params) Int32(key string) int32 {
	return int32(p.Int64(key))
}

func (p Params) Bool(key string) bool {
	s := p.Get(key)
	for _, t := range truthy {
		if s == t {
			return true
		}
	}
	return false
}

func (p Params) Map() (result map[string]string) {
	result = make(map[string]string)
	for k, v := range p {
		if k != "" {
			if len(v) == 1 && v[0] != "" {
				result[k] = v[0]
			}
		}
	}
	return
}

func (p Params) Int64s(key string) []int64 {
	ns := make([]int64, 0)
	val := p.Get(key)
	ss := strings.Split(val, ",")
	for _, s := range ss {
		n, err := strconv.ParseInt(s, 10, 0)
		if err == nil {
			ns = append(ns, n)
		}
	}

	return ns
}
