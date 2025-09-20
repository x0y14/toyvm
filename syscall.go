package toyvm

type SystemCall int

func (s SystemCall) Value() int {
	return int(s)
}
func (s SystemCall) String() string {
	switch s {
	case Write:
		return "write"
	default:
		return ""
	}
}

const (
	Write SystemCall = iota
)
