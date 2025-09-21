package toyvm

import (
	"fmt"
	"strconv"
)

type MemoryOffset int

func (m MemoryOffset) Value() int {
	return int(m)
}
func (m MemoryOffset) String() string {
	return strconv.Itoa(m.Value())
}

type Memory []Object

func NewMemory(size int) *Memory {
	mem := make(Memory, size)
	return &mem
}

func (m *Memory) Set(offset MemoryOffset, obj Object) error {
	if 0 <= offset.Value() && offset.Value() < len(*m) {
		(*m)[offset.Value()] = obj
		return nil
	}
	return fmt.Errorf("offset must be 0< = x < %d", len(*m))
}

func (m *Memory) Get(offset MemoryOffset) (Object, error) {
	if 0 <= offset.Value() && offset.Value() < len(*m) {
		return (*m)[offset.Value()], nil
	}
	return nil, fmt.Errorf("out of range: %d in %d", offset, len(*m))
}

func (m *Memory) Delete(offset MemoryOffset) {
	(*m)[offset.Value()] = nil
}

func (m *Memory) IsEmpty(offset MemoryOffset) bool {
	return (*m)[offset.Value()] == nil
}
