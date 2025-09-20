package toyvm

import (
	"fmt"
	"strconv"
)

type ProgramAbsoluteOffset int

func (p ProgramAbsoluteOffset) Value() int {
	return int(p)
}
func (p ProgramAbsoluteOffset) String() string {
	return fmt.Sprintf("pc(%d)", p.Value())
}

type StackRelativeOffset struct {
	target           Register
	relativeDistance int
}

func NewBPOffset(relativeDistance int) *StackRelativeOffset {
	return &StackRelativeOffset{BasePointer, relativeDistance}
}

func (s StackRelativeOffset) Value() int {
	return s.relativeDistance
}
func (s StackRelativeOffset) String() string {
	str := "["
	str += s.target.String()
	if 0 <= s.Value() {
		str += "+"
	}
	str += strconv.Itoa(s.Value())
	str += "]"
	return fmt.Sprintf("%s(%d)", s.target.String(), s.Value())
}
