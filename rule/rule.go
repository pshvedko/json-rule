package rule

import (
	"fmt"
	"github.com/pshvedko/json-rule/jsonpath"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
)

type Builder struct {
	strings.Builder
	Group bool
	Paths map[string][]string
	Cache [][]string
	Names []string
}

func (b *Builder) Print(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(b, format, a...)
}

func (b *Builder) Contains(e string) bool {
	for _, n := range b.Names {
		if n == e {
			return true
		}
	}
	return false
}

func (b *Builder) Variable(e, f string) string {
	p := strings.Split(e, ".")
	q := strings.Split(f, ".")
	p = append(p, q...)
	v := fmt.Sprint("x", b.Count(p))
	b.Paths[v] = p
	return v
}

func (b *Builder) Count(p []string) int {
	var i int
	for i < len(b.Cache) {
		if reflect.DeepEqual(b.Cache[i], p) {
			return i
		}
		i++
	}
	b.Cache = append(b.Cache, p)
	return i
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
		if !b.Contains(o.Event) {
			return io.EOF
		}
		_, err = b.Print("%s%s%s", p, b.Variable(o.Event, o.Field), e)
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

func (e Expression) Build(events []string) (string, map[string][]string, error) {
	b := Builder{
		Builder: strings.Builder{},
		Group:   false,
		Paths:   map[string][]string{},
		Names:   events,
	}
	for _, o := range e {
		if err := o.Print(&b); err != nil {
			return "", nil, err
		}
	}
	return b.String(), b.Paths, nil
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
	b, p, err := r.Body.Expression.Build(r.BasicEvents)
	if err != nil {
		return nil, err
	}
	var q *govaluate.EvaluableExpression
	q, err = govaluate.NewEvaluableExpressionWithFunctions(b, internalFunctions)
	if err != nil {
		return nil, err
	}
	return func(j interface{}) (interface{}, error) {
		return q.Eval(jsonpath.NewGetterWithPreparedPath(j, p))
	}, nil
}
