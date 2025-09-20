package toyvm

import (
	"fmt"
	"strconv"
)

type Label int

func (l Label) Value() int {
	return int(l)
}
func (l Label) String() string {
	return strconv.Itoa(l.Value())
}

type DefLabel int

func (d DefLabel) Value() int {
	return int(d)
}
func (d DefLabel) String() string {
	return fmt.Sprintf("%d:", d.Value())
}
