package path

import (
	"errors"
	"strconv"
	"strings"
)

type GetterFunc func(from interface{}) (interface{}, error)

func (f GetterFunc) Get(from interface{}) (interface{}, error) {
	return f(from)
}

type Getter interface {
	Get(from interface{}) (interface{}, error)
}

func NewGetter(path string) Getter {
	keys := strings.Split(path, ".")
	return GetterFunc(func(from interface{}) (interface{}, error) {
		return Get(from, keys)
	})
}

var ErrInvalidKey = errors.New("invalid key")

func Get(j interface{}, k []string) (interface{}, error) {
	if len(k) > 0 {
		switch m := j.(type) {
		case map[string]interface{}:
			v, ok := m[k[0]]
			if !ok {
				break
			}
			return Get(v, k[1:])
		case []interface{}:
			if k[0] == "#" {
				a := make([]interface{}, 0, len(m))
				for _, v := range m {
					v, err := Get(v, k[1:])
					if err != nil {
						return nil, err
					}
					a = append(a, v)
				}
				return a, nil
			}
			i, err := strconv.Atoi(k[0])
			if err != nil || i < 0 || i >= len(m) {
				break
			}
			return Get(m[i], k[1:])
		}
		return nil, ErrInvalidKey
	}
	return j, nil
}
