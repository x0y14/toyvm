package toyvm

import (
	"fmt"
)

type SymbolTable map[Label]ProgramAbsoluteOffset

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{}
}

func (s *SymbolTable) Get(label Label) (ProgramAbsoluteOffset, error) {
	v, ok := (*s)[label]
	if !ok {
		return v, fmt.Errorf("failed to get symbol: not registered: %v", label)
	}
	return v, nil
}

func (s *SymbolTable) Set(label Label, offset ProgramAbsoluteOffset) error {
	_, err := s.Get(label)
	if err == nil {
		return fmt.Errorf("failed to set symbol: already registered: %v", label)
	}
	(*s)[label] = offset
	return nil
}

func (s *SymbolTable) Delete(label Label) {
	delete(*s, label)
}
