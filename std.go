package toyvm

type StandardIO int

func (s StandardIO) Value() int {
	return int(s)
}
func (s StandardIO) String() string {
	switch s {
	case StdIn:
		return "stdin"
	case StdOut:
		return "stdout"
	case StdErr:
		return "stderr"
	default:
		return ""
	}
}

const (
	StdIn StandardIO = iota
	StdOut
	StdErr
)
