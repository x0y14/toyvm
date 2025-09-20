package toyvm

type Register int

func (r Register) Value() int {
	return int(r)
}
func (r Register) String() string {
	var kinds = [...]string{
		ProgramCounter: "pc",
		BasePointer:    "bp",
		StackPointer:   "sp",
		ZeroFlag:       "zf",
		ExitFlag:       "ef",
		General1:       "g1",
		General2:       "g2",
		Temporal1:      "t1",
		R0:             "r0",
		R1:             "r1",
		R2:             "r2",
		R3:             "r3",
		R4:             "r4",
		R5:             "r5",
		R6:             "r6",
		R7:             "r7",
		R8:             "r8",
		R9:             "r9",
		R10:            "r10",
		R11:            "r11",
		R12:            "r12",
		ACM1:           "acm1",
		ACM2:           "acm2",
		_reg_end:       "",
	}
	return kinds[r]
}

const (
	ProgramCounter Register = iota
	BasePointer
	StackPointer
	ZeroFlag
	ExitFlag
	General1
	General2
	Temporal1

	R0
	R1
	R2
	R3
	R4
	R5
	R6
	R7
	R8
	R9
	R10 // 計算結果が入る
	R11 // 計算結果が入る
	R12

	ACM1
	ACM2

	_reg_end
)

func NewRegisterSet() *[]Object {
	rSet := make([]Object, _reg_end-1)
	return &rSet
}
