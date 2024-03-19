package datastructure

import (
	"container/list"
	"reflect"
	"testing"
)

func TestQueue_Pop(t *testing.T) {
	type fields struct {
		l *list.List
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{"test1", fields{l: list.New()}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Queue{
				l: tt.fields.l,
			}
			if got := q.Pop(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueue_Push(t *testing.T) {
	type fields struct {
		l *list.List
	}
	type args struct {
		v interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"test1", fields{l: list.New()}, args{v: "test"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Queue{
				l: tt.fields.l,
			}
			q.Push(tt.args.v)
		})
	}
}
