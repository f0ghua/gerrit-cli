package query

import (
	"fmt"
	"strings"
)

type Query struct {
	predicates []string
	limit      int
}

func New() *Query {
	return &Query{}
}

func (q *Query) Owner(owner string) *Query {
	q.predicates = append(q.predicates, "owner:"+owner)
	return q
}

func (q *Query) Reviewer(reviewer string) *Query {
	q.predicates = append(q.predicates, "reviewer:"+reviewer)
	return q
}

func (q *Query) Status(status string) *Query {
	q.predicates = append(q.predicates, "status:"+status)
	return q
}

func (q *Query) Project(project string) *Query {
	q.predicates = append(q.predicates, "project:"+project)
	return q
}

func (q *Query) Branch(branch string) *Query {
	q.predicates = append(q.predicates, "branch:"+branch)
	return q
}

func (q *Query) Topic(topic string) *Query {
	q.predicates = append(q.predicates, "topic:"+topic)
	return q
}

func (q *Query) Label(label string) *Query {
	q.predicates = append(q.predicates, "label:"+label)
	return q
}

func (q *Query) After(date string) *Query {
	q.predicates = append(q.predicates, "after:"+date)
	return q
}

func (q *Query) Before(date string) *Query {
	q.predicates = append(q.predicates, "before:"+date)
	return q
}

func (q *Query) Is(val string) *Query {
	q.predicates = append(q.predicates, "is:"+val)
	return q
}

func (q *Query) Has(val string) *Query {
	q.predicates = append(q.predicates, "has:"+val)
	return q
}

func (q *Query) Raw(raw string) *Query {
	q.predicates = append(q.predicates, raw)
	return q
}

func (q *Query) Limit(n int) *Query {
	q.limit = n
	return q
}

func (q *Query) GetLimit() int {
	if q.limit <= 0 {
		return 25
	}
	return q.limit
}

func (q *Query) String() string {
	return strings.Join(q.predicates, " ")
}

func (q *Query) GoString() string {
	return fmt.Sprintf("Query{%s, limit=%d}", q.String(), q.GetLimit())
}
