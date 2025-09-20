package toyvm

type Opcode int

func (o Opcode) Value() int {
	return int(o)
}

func (o Opcode) String() string {
	var kinds = [...]string{
		Nop: "Nop",

		Exit: "Exit",

		Mov:  "Mov",
		Push: "Push",
		Pop:  "Pop",

		Call: "Call",
		Ret:  "Ret",

		Add: "Add",
		Sub: "Sub",

		Jmp: "Jmp",
		Je:  "Je",
		Jne: "Jne",

		Eq:      "Eq",
		Ne:      "Ne",
		Lt:      "Lt",
		Le:      "Le",
		Syscall: "Syscall",
	}
	return kinds[o]
}

const (
	Nop Opcode = iota

	Exit

	Mov
	Push
	Pop

	Call
	Ret

	Add
	Sub

	Jmp
	Je
	Jne

	Eq
	Ne
	Lt
	Le

	Syscall
)

func Operand(op Opcode) int {
	switch op {
	case Nop, Exit, Ret:
		return 0
	case Push, Pop, Call, Jmp, Je, Jne:
		return 1
	case Mov, Add, Sub, Eq, Ne, Lt, Le:
		return 2
	case Syscall:
		return 3
	default:
		return 0
	}
}
