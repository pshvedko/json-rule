package rule

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/pshvedko/json-rule/jsonpath"
)

type Quoter struct {
	*Builder
}

func (r Quoter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		switch b {
		case '.', '#', ':':
			err = r.WriteByte('\\')
			if err != nil {
				return
			}
			n++
		}
		err = r.WriteByte(b)
		if err != nil {
			return
		}
		n++
	}
	return
}

func (r Quoter) Print(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(r, format, a...)
}

type Builder struct {
	strings.Builder
	Group bool
}

func (b *Builder) Print(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(b, format, a...)
}

func (b *Builder) Quoter() Quoter {
	return Quoter{
		Builder: b,
	}
}

type Point struct {
	Token string `json:"token,omitempty"`
	Type  string `json:"type,omitempty"`
}

type Points []Point

type ExitPoints struct {
	IsEveryCondition bool   `json:"is_every_condition,omitempty"`
	Points           Points `json:"points,omitempty"`
}

type Action struct {
	EventType string `json:"event_type,omitempty"`
	Name      string `json:"name,omitempty"`
}

type Operand struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
	Event string `json:"event,omitempty"`
	Field string `json:"field,omitempty"`
}

func (o Operand) Print(b *Builder) (err error) {
	var p, e, q string
	switch o.Type {
	//case "int":
	//	p, e = o.Type+"(", ")"
	case "string":
		q = "'"
	}
	if o.Value != "" {
		_, err = b.Print("%s%s%s%s%s", p, q, o.Value, q, e)
	} else {
		_, err = b.Quoter().Print("%s%s::%s%s", p, o.Event, o.Field, e)
	}
	return
}

type Operation struct {
	Group    bool    `json:"group,omitempty"`
	Left     Operand `json:"left"`
	Action   string  `json:"action,omitempty"`
	Right    Operand `json:"right"`
	Operator string  `json:"operator,omitempty"`
}

func (o Operation) Print(b *Builder) (err error) {
	defer func() {
		if err == nil {
			_, err = b.Print(" %s ", o.Operator)
		}
	}()
	if b.Group != o.Group {
		switch o.Group {
		case true:
			err = b.WriteByte('(')
			if err != nil {
				return
			}
		case false:
			defer func() {
				if err == nil {
					err = b.WriteByte(')')
				}
			}()
		}
		b.Group = o.Group
	}
	err = o.Left.Print(b)
	if err != nil {
		return
	}
	_, err = b.Print(" %s ", o.Action)
	if err != nil {
		return
	}
	return o.Right.Print(b)
}

type Expression []Operation

func (e Expression) Build() (string, error) {
	var b Builder
	for _, o := range e {
		if err := o.Print(&b); err != nil {
			return "", nil
		}
	}
	return b.String(), nil
}

type Actions []Action

type Body struct {
	Actions    Actions    `json:"actions,omitempty"`
	Expression Expression `json:"expression,omitempty"`
}

type Rule struct {
	BasicEvents      []string   `json:"basic_events,omitempty"`
	Body             Body       `json:"body"`
	CreatedDate      time.Time  `json:"created_date"`
	Creator          string     `json:"creator,omitempty"`
	Description      string     `json:"description,omitempty"`
	ExitPoints       ExitPoints `json:"exit_points"`
	Id               string     `json:"id,omitempty"`
	Initiator        string     `json:"initiator,omitempty"`
	KeyField         string     `json:"key_field,omitempty"`
	ModificationDate time.Time  `json:"modification_date"`
	Name             string     `json:"name,omitempty"`
	Status           string     `json:"status,omitempty"`
	Type             string     `json:"type,omitempty"`
	Weight           int        `json:"weight,omitempty"`
}

type Getter func(x string) (interface{}, error)

func (g Getter) Get(x string) (interface{}, error) {
	return g(x)
}

type Key struct {
	e int
	k []string
}

type Condition struct {
	g *govaluate.EvaluableExpression
	f map[string]map[string]Key
}

func (c Condition) Exec(j ...interface{}) (interface{}, error) {
	return c.g.Eval(Getter(func(x string) (interface{}, error) {
		if u := strings.SplitN(x, "::", 2); 2 == len(u) {
			if f, ok := c.f[u[0]]; ok {
				if k, ok := f[u[1]]; ok {
					return jsonpath.Get(j[k.e], k.k)
				}
			}
		}
		return nil, os.ErrInvalid
	}))
}
func (c Condition) String() string {
	return c.g.String()
}

func (c Condition) Variables() []string {
	return c.g.Vars()
}

var internalFunctions = map[string]govaluate.ExpressionFunction{
	"int": func(a ...interface{}) (interface{}, error) {
		return a[0], nil
	},
}

func (r Rule) Condition() (c Condition, err error) {
	var x string
	x, err = r.Body.Expression.Build()
	if err != nil {
		return
	}
	c.g, err = govaluate.NewEvaluableExpressionWithFunctions(x, internalFunctions)
	if err != nil {
		return
	}
	c.f = map[string]map[string]Key{}
	for _, v := range c.g.Vars() {
		for i, e := range r.BasicEvents {
			if l := len(e); l > 0 && v[:l] == e && v[l] == ':' && v[1+l] == ':' {
				if _, ok := c.f[e]; !ok {
					c.f[e] = map[string]Key{}
				}
				if _, ok := c.f[e][v[2+l:]]; !ok {
					c.f[e][v[2+l:]] = Key{
						e: i,
						k: strings.Split(v[2+l:], "."),
					}
				}
			}
		}
	}
	return
}
