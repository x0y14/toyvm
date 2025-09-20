package toyvm

import (
	"fmt"
	"strconv"
)

type Object interface {
	Value() int
	String() string
}

type Integer int

func (i Integer) Value() int {
	return int(i)
}
func (i Integer) String() string {
	return strconv.Itoa(i.Value())
}

type Character int

func (c Character) Value() int {
	return int(c)
}
func (c Character) String() string {
	return string(rune(c.Value()))
}

type Bool bool

func (b Bool) Value() int {
	if b == true {
		return 1
	}
	return 0
}
func (b Bool) String() string {
	if b == true {
		return "true"
	}
	return "false"
}

const (
	True  Bool = true
	False Bool = false
)

type Null struct{}

func (n Null) Value() int {
	return 0
}
func (n Null) String() string {
	return "null"
}

type List int

func (l List) Value() int {
	return int(l)
}
func (l List) String() string {
	return fmt.Sprintf("list(%d)", l.Value())
}
