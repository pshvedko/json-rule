package rule

import (
	"fmt"
	"github.com/pshvedko/json-rule/jsonpath"
	"os"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
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
		_, err = b.Quoter().Print("%s%s.%s%s", p, o.Event, o.Field, e)
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

var internalFunctions = map[string]govaluate.ExpressionFunction{
	"int": func(a ...interface{}) (interface{}, error) {
		return a[0], nil
	},
}

type Condition func(j interface{}) (interface{}, error)

func (c Condition) Evaluate(j interface{}) (interface{}, error) {
	if c != nil {
		return c(j)
	}
	return nil, os.ErrInvalid
}

func (r Rule) Condition() (Condition, error) {
	x, err := r.Body.Expression.Build()
	if err != nil {
		return nil, err
	}
	var q *govaluate.EvaluableExpression
	q, err = govaluate.NewEvaluableExpressionWithFunctions(x, internalFunctions)
	if err != nil {
		return nil, err
	}
	p := map[string][]string{}
	n := 0
	for i, v := range q.Vars() {
		for _, e := range r.BasicEvents {
			if l := len(e); l > 0 && e == v[:l] && v[l] == '.' {
				if _, ok := p[v]; !ok {
					p[v] = strings.Split(v, ".")
				}
				n++
				break
			}
		}
		if i == n {
			return nil, os.ErrInvalid
		}
	}
	return func(j interface{}) (interface{}, error) {
		return q.Eval(jsonpath.NewGetterWithPreparedPath(j, p))
	}, nil
}
