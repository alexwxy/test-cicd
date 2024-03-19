package datastructure

import "container/list"

type Queue struct {
	l *list.List
}

func NewQueue() *Queue {
	return &Queue{l: list.New()}
}

func (q *Queue) Push(v interface{}) {
	q.l.PushBack(v)
}

func (q *Queue) Pop() interface{} {
	e := q.l.Front()
	if e != nil {
		q.l.Remove(e)
		return e.Value
	}
	return nil
}

func (q *Queue) Len() int {
	return q.l.Len()
}

func (q *Queue) IsEmpty() bool {
	return q.l.Len() == 0
}
