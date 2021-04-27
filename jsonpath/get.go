package jsonpath

import (
	"os"
	"strconv"
	"strings"
)

type Getter func(x string) (interface{}, error)

func (g Getter) Get(x string) (interface{}, error) {
	return g(x)
}

func NewGetter(j interface{}) Getter {
	return func(x string) (interface{}, error) {
		return Get(j, strings.Split(x, "."))
	}
}

func NewGetterWithPreparedPath(j interface{}, p map[string][]string) Getter {
	return func(x string) (interface{}, error) {
		if v, ok := p[x]; ok {
			return Get(j, v)
		}
		return nil, os.ErrInvalid
	}
}

// Get looks up value in JSON specified by keys. Any keys can be specified by #
func Get(j interface{}, k []string) (interface{}, error) {
	if len(k) > 0 {
		switch m := j.(type) {
		case map[string]interface{}:
			if k[0] == "#" {
				a := make(map[string]interface{})
				for i, v := range m {
					v, err := Get(v, k[1:])
					if err != nil {
						continue
					}
					a[i] = v
				}
				return a, nil
			}
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
						continue
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
		return nil, os.ErrInvalid
	}
	return j, nil
}
