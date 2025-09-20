package toyvm

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRuntime(t *testing.T) {
	// スタックは自動で設定される
	reg := *NewRegisterSet()
	reg[StackPointer] = Integer(1 - 1)
	reg[ProgramCounter] = Integer(0)
	reg[BasePointer] = Integer(0)
	assert.Equal(t, &Runtime{
		program: nil,
		sym:     *NewSymbolTable(),
		reg:     reg,
		stack:   make([]Object, 1),
		mem:     make(Memory, 1),
	}, NewRuntime(1, 1))
	//
	reg = *NewRegisterSet()
	reg[StackPointer] = Integer(15 - 1)
	reg[ProgramCounter] = Integer(0)
	reg[BasePointer] = Integer(0)
	assert.Equal(t, &Runtime{
		program: nil,
		sym:     *NewSymbolTable(),
		reg:     reg,
		stack:   make([]Object, 15),
		mem:     make(Memory, 15),
	}, NewRuntime(15, 15))
}

func TestRuntime_CollectLabels(t *testing.T) {
	rt := NewRuntime(1, 1)
	rt.Load(Program{
		// main:
		//   ret
		DefLabel(0),
		Ret,
	})
	assert.Nil(t, rt.CollectLabels())
	// 擬似プロセスコードが挿入されるのでコードが追加される
	assert.Equal(t, Program{
		DefLabel(-1),   // ここから->
		Call, Label(0), //
		Exit,        // <- ここまで追加
		DefLabel(0), // ここ以降はユーザーが読み込ませたプログラム
		Ret,
	}, rt.program)
	assert.Equal(t, ProgramAbsoluteOffset(0), rt.sym[-1]) // root label, DefLabel(-1)の配列での位置
	assert.Equal(t, ProgramAbsoluteOffset(4), rt.sym[0])  // main label, DefLabel(0)の配列での位置
}

func TestRuntime_Run_Exit(t *testing.T) {
	// Mainなし
	rt := NewRuntime(1, 1)
	rt.program = Program{
		Exit,
	}
	rt.sym[Label(-1)] = ProgramAbsoluteOffset(0) // RunでcallされるものをCollectLabelsの代わりにセットしてあげる
	err := rt.Run()
	assert.Nil(t, err)
	// Mainあり
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
}

func TestRuntime_Run_Call(t *testing.T) {
	rt := NewRuntime(10, 10)
	rt.Load(Program{
		DefLabel(5),
		Ret,

		DefLabel(4),
		Call, Label(5),
		Ret,

		DefLabel(3),
		Call, Label(4),
		Ret,

		DefLabel(2),
		Call, Label(3),
		Ret,

		DefLabel(1),
		Call, Label(2),
		Ret,

		DefLabel(0),
		Call, Label(1),
		Ret,
	})
	err := rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
}

func TestRuntime_Run_Mov(t *testing.T) {
	rt := NewRuntime(10, 10)
	rt.Load(Program{
		DefLabel(5),
		Mov, General1, Integer(5),
		Ret,

		DefLabel(4),
		Mov, General1, Integer(4),
		Call, Label(5),
		Ret,

		DefLabel(3),
		Mov, General1, Integer(3),
		Call, Label(4),
		Ret,

		DefLabel(2),
		Mov, General1, Integer(2),
		Call, Label(3),
		Ret,

		DefLabel(1),
		Mov, General1, Integer(1),
		Call, Label(2),
		Ret,

		DefLabel(0),
		Call, Label(1),
		Ret,
	})
	err := rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, Integer(5), rt.reg[General1])
}

func TestRuntime_Run_Add(t *testing.T) {
	rt := NewRuntime(10, 10)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(1),
		Add, General1, Integer(4),
		Ret,
	})
	err := rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, Integer(5), rt.reg[General1])
}

func TestRuntime_Run_Sub(t *testing.T) {
	rt := NewRuntime(10, 10)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(5),
		Sub, General1, Integer(3),
		Ret,
	})
	err := rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, Integer(2), rt.reg[General1])
}

func TestRuntime_Run_Push(t *testing.T) {
	rt := NewRuntime(10, 10)
	rt.program = Program{
		DefLabel(-1),
		Push, Integer(1),
		Push, Integer(2),
		Push, Integer(3),
		Exit,
	}
	rt.sym[Label(-1)] = ProgramAbsoluteOffset(0)
	err := rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, Integer(3), rt.stack[6]) // どう動いてるかよくわからん
	assert.Equal(t, Integer(3), rt.pop())
	assert.Equal(t, Integer(2), rt.stack[7])
	assert.Equal(t, Integer(2), rt.pop())
	assert.Equal(t, Integer(1), rt.stack[8])
	assert.Equal(t, Integer(1), rt.pop())
}
func TestRuntime_Run_Pop(t *testing.T) {
	rt := NewRuntime(10, 10)
	rt.Load(Program{
		DefLabel(0),
		Push, Integer(1),
		Push, Integer(2),
		Pop, General1, // <- 2
		Pop, General2, // <- 1
		Ret,
	})
	err := rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, Integer(2), rt.reg[General1])
	assert.Equal(t, Integer(1), rt.reg[General2])
}

func TestRuntime_Run_Eq(t *testing.T) {
	// reg(int) == int, want true
	rt := NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(1),
		Eq, General1, Integer(1),
		Ret,
	})
	err := rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, True, rt.reg[ZeroFlag])
	// reg(int) == int, want false
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(1),
		Eq, General1, Integer(3),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, False, rt.reg[ZeroFlag])
	// reg(true) == int, want false
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, True,
		Eq, General1, Integer(1),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, False, rt.reg[ZeroFlag])
	// reg(char) == int, want false
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Character(1), // 数値が同じでも型が違うとFalseになることを確認
		Eq, General1, Integer(1),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, False, rt.reg[ZeroFlag])
}
func TestRuntime_Run_Ne(t *testing.T) {
	// reg(int) != int, want false
	rt := NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(1),
		Ne, General1, Integer(1),
		Ret,
	})
	err := rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, False, rt.reg[ZeroFlag])
	// reg(int) != int, want true
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(1),
		Ne, General1, Integer(3),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, True, rt.reg[ZeroFlag])
	// reg(true) != int, want true
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, True,
		Ne, General1, Integer(1),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, True, rt.reg[ZeroFlag])
	// reg(char) != int, want true
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Character(1), // 数値が同じでも型が違うとTrueになることを確認
		Ne, General1, Integer(1),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, True, rt.reg[ZeroFlag])
}

func TestRuntime_Run_Lt(t *testing.T) {
	// 1 < 1, false
	rt := NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(1),
		Lt, General1, Integer(1),
		Ret,
	})
	err := rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, False, rt.reg[ZeroFlag])
	// 1 < 2, true
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(1),
		Lt, General1, Integer(2),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, True, rt.reg[ZeroFlag])
	// 2 < 1, false
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(2),
		Lt, General1, Integer(1),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, False, rt.reg[ZeroFlag])
	// char(1) < int(2), true
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Character(1),
		Lt, General1, Integer(2),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, True, rt.reg[ZeroFlag]) // 大きさ比較なので型が違うものも許可してる.
}

func TestRuntime_Run_Le(t *testing.T) {
	// 1 <= 1, true
	rt := NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(1),
		Le, General1, Integer(1),
		Ret,
	})
	err := rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, True, rt.reg[ZeroFlag])
	// 1 <= 2, true
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(1),
		Le, General1, Integer(2),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, True, rt.reg[ZeroFlag])
	// 2 <= 1, false
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(2),
		Le, General1, Integer(1),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, False, rt.reg[ZeroFlag])
	// char(1) <= int(2), true
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Character(1),
		Le, General1, Integer(2),
		Ret,
	})
	err = rt.CollectLabels()
	assert.Nil(t, err)
	err = rt.Run()
	assert.Nil(t, err)
	assert.Equal(t, True, rt.reg[ZeroFlag]) // 大きさ比較なので型が違うものも許可してる.
}

func TestRuntime_Run_Jmp(t *testing.T) {
	rt := NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(0), // g1 = 0
		Jmp, Label(1),
		Add, General1, Integer(1), // g1 += 1, スキップされるはず
		DefLabel(1),
		Add, General1, Integer(2), // g1 += 2, これだけ実行されるはず
		Ret,
	})
	assert.Nil(t, rt.CollectLabels())
	assert.Nil(t, rt.Run())
	assert.Equal(t, Integer(2), rt.reg[General1])
	// Jmpを抜くと g1==3 になることを確認
	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(0), // g1 = 0
		//Jmp, Label(1),
		Add, General1, Integer(1), // g1 += 1
		DefLabel(1),
		Add, General1, Integer(2), // g1 += 2
		Ret,
	})
	assert.Nil(t, rt.CollectLabels())
	assert.Nil(t, rt.Run())
	assert.Equal(t, Integer(3), rt.reg[General1])
}
func TestRuntime_Run_Je(t *testing.T) {
	rt := NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(0), // g1 = 0
		Eq, Integer(0), Integer(0), // 0 == 0?
		Je, Label(1), // if zf==1, goto l_1
		Add, General1, Integer(1), // g1 += 1, skip
		DefLabel(1),
		Add, General1, Integer(2), // g1 += 2, do
		Ret,
	})
	assert.Nil(t, rt.CollectLabels())
	assert.Nil(t, rt.Run())
	assert.Equal(t, Integer(2), rt.reg[General1])

	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(0), // g1 = 0
		Eq, Integer(0), Integer(1), // 0 == 1?
		Je, Label(1), // if zf==1, goto l_1
		Add, General1, Integer(1), // g1 += 1, do
		DefLabel(1),
		Add, General1, Integer(2), // g1 += 2, do
		Ret,
	})
	assert.Nil(t, rt.CollectLabels())
	assert.Nil(t, rt.Run())
	assert.Equal(t, Integer(3), rt.reg[General1])
}
func TestRuntime_Run_Jne(t *testing.T) {
	rt := NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(0), // g1 = 0
		Eq, Integer(0), Integer(0), // 0 == 0?
		Jne, Label(1), // if zf==0, goto l_1
		Add, General1, Integer(1), // g1 += 1, skip
		DefLabel(1),
		Add, General1, Integer(2), // g1 += 2, do
		Ret,
	})
	assert.Nil(t, rt.CollectLabels())
	assert.Nil(t, rt.Run())
	assert.Equal(t, Integer(3), rt.reg[General1])

	rt = NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Mov, General1, Integer(0), // g1 = 0
		Eq, Integer(0), Integer(1), // 0 == 1?
		Jne, Label(1), // if zf==0, goto l_1
		Add, General1, Integer(1), // g1 += 1, do
		DefLabel(1),
		Add, General1, Integer(2), // g1 += 2, do
		Ret,
	})
	assert.Nil(t, rt.CollectLabels())
	assert.Nil(t, rt.Run())
	assert.Equal(t, Integer(2), rt.reg[General1])
}

func TestRuntime_Run_Syscall_Write(t *testing.T) {
	tmpStdout := os.Stdout // 標準出力を元に戻せるように保存
	r, w, _ := os.Pipe()
	os.Stdout = w // 標準出力の書き込み先を変更

	rt := NewRuntime(2, 1)
	rt.Load(Program{
		DefLabel(0),
		Syscall, Write, StdOut, Character('h'),
		Syscall, Write, StdOut, Character('e'),
		Syscall, Write, StdOut, Character('l'),
		Syscall, Write, StdOut, Character('l'),
		Syscall, Write, StdOut, Character('o'),
		Syscall, Write, StdOut, Character(','),
		Syscall, Write, StdOut, Character('w'),
		Syscall, Write, StdOut, Character('o'),
		Syscall, Write, StdOut, Character('r'),
		Syscall, Write, StdOut, Character('l'),
		Syscall, Write, StdOut, Character('d'),
		Syscall, Write, StdOut, Character('!'),
		Ret,
	})
	assert.Nil(t, rt.CollectLabels())
	assert.Nil(t, rt.Run())

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	s := strings.TrimRight(buf.String(), "") // バッファーから文字列へ変換
	os.Stdout = tmpStdout
	assert.Equal(t, "hello,world!", s)
}

func TestRuntime_Run_FizzBuzz(t *testing.T) {
	tmpStdout := os.Stdout // 標準出力を元に戻せるように保存
	r, w, _ := os.Pipe()
	os.Stdout = w // 標準出力の書き込み先を変更

	rt := NewRuntime(1000, 10)
	rt.Load(Program{
		//fn check_x(n int, x int) bool {
		//	n = n - x
		//	if n == 0 {
		//		return true
		//	} else if n < 0 {
		//		return false
		//	}
		//	return check_x(n, x)
		//}

		// check_x:
		DefLabel(1),
		// # 関数の初期設定 #
		// ## 現状復帰のための保存 ##
		Push, BasePointer,
		Mov, BasePointer, StackPointer,
		// ## 引数を含む変数領域の確保 ##
		Push, Integer(2), // xとnの2つ
		Pop, R1, // 一旦レジスタに入れてから引き算する
		Sub, StackPointer, R1, // SPをずらす
		// ## 引数と変数の結び付け ##
		// Arg1
		Mov, StackRelativeOffset{BasePointer, -1}, StackRelativeOffset{BasePointer, +2},
		// Arg2
		Mov, StackRelativeOffset{BasePointer, -2}, StackRelativeOffset{BasePointer, +3},
		//
		// # n -= xをする #
		// ## n - x ##
		Push, StackRelativeOffset{BasePointer, -1}, // 左辺n
		Push, StackRelativeOffset{BasePointer, -2}, // 右辺x
		// < sub-beg >
		Pop, R2, // 右辺の取り出し
		Pop, R1, // 左辺の取り出し
		Sub, R1, R2, // R1 -= R2
		Push, R1,
		// < sub-end >
		// ## nに計算結果を代入する ##
		// < assign-beg >
		Pop, R1, // 直前の結果の取り出し
		Mov, StackRelativeOffset{BasePointer, -1}, R1, // n = ${直前}
		// < assign-end >
		//
		// # if n == 0 {} をする #
		// < if-beg >
		// ## n == 0 ##
		Push, StackRelativeOffset{BasePointer, -1}, // 左辺n
		Push, Integer(0), // 右辺
		// < eq-beg >
		Pop, R2, // 右辺の取り出し
		Pop, R1, // 左辺の取り出し
		Eq, R1, R2, // r1 == r2
		Push, ZeroFlag,
		// < eq-end >
		// ## 条件分岐 ##
		Pop, Temporal1, // 直前の結果いらないので捨てる
		Je, Label(2), // check_x_if_1_if, (n == 0) == trueだったらifブロックへ
		Jmp, Label(3), // check_x_if_1_else, (n == 0) == falseだったらelseブロックへ
		// check_x_if_1_if:
		DefLabel(2),
		// # if n == 0 {}のリターン処理 #
		// < return-beg >
		// ## 戻り値の設定 ##
		Mov, ACM1, True,
		// # 関数の終了処理 #
		Mov, StackPointer, BasePointer,
		Pop, BasePointer,
		Ret,
		// < return-end >
		Jmp, Label(4), // check_x_if_1_end, ifの終了へ
		// check_x_if_1_else:
		DefLabel(3),
		// 何もしない
		Jmp, Label(4), // check_x_if_1_end, ifの終了へ
		// check_x_if_1_end:
		DefLabel(4),
		// < if-end >
		//
		// # if n < 0 {} をする #
		// < if-beg >
		// ## n < 0 ##
		Push, StackRelativeOffset{BasePointer, -1}, // 左辺n
		Push, Integer(0),
		// < lt-beg >
		Pop, R2, // 右辺の取り出し
		Pop, R1, // 左辺の取り出し
		Lt, R1, R2, // r1 < r2
		Push, ZeroFlag, // 結果
		// < lt-end >
		// ## 条件分岐 ##
		Pop, Temporal1,
		Je, Label(5), // check_x_if_2_if, (n < 0) == trueならifブロックへ
		Jmp, Label(6), // check_x_if_2_else, (n < 0) == falseならelseブロックへ
		// check_x_if_2_if:
		DefLabel(5),
		// # if n < 0 {} のリターン処理 #
		// < return-beg >
		// ## 戻り値の設定 ##
		Mov, ACM1, False,
		// # 関数の終了処理 #
		Mov, StackPointer, BasePointer,
		Pop, BasePointer,
		Ret,
		// < return-end >
		Jmp, Label(7), // check_x_if_2_end, ifの終了へ
		// check_x_if_2_else:
		DefLabel(6),
		// 何もしない
		Jmp, Label(7), // check_x_if_2_end, ifの終了へ
		// check_x_if_2_end:
		DefLabel(7),
		// < if-end >
		//
		// どれのifにも引っ掛かんなかった場合のreturn
		// < return-beg >
		// # 戻り値の用意 #
		// check_x(n,x)
		Push, StackRelativeOffset{BasePointer, -2}, // xをarg2として
		Push, StackRelativeOffset{BasePointer, -1}, // nをarg1として
		// < call-beg >
		Call, Label(1), // check_x
		Push, Integer(2), // 引数分spを戻す
		Pop, R1,
		Add, StackPointer, R1,
		Push, ACM1, // 結果を取り出す
		// < call-end >
		// # 戻り値の格納
		Pop, R1,
		Mov, ACM1, R1,
		// < return-end >
		//
		// # 関数の終了処理 #
		Mov, StackPointer, BasePointer,
		Pop, BasePointer,
		Ret,

		// ---

		// fn main() {
		//     i := 1
		//     for {
		//         if check(i, 15) { println(i, "fizzbuzz") }
		//         else if check(i, 5) { println(i, "buzz") }
		//         else if check(i, 3) { println(i, "fizz") }
		//         else { println(i) }
		//         i ++
		//         if i  == 101 { break }
		//     }
		//     return
		// }

		// main:
		DefLabel(0),
		// # 関数の初期設定 #
		// ## 現状復帰のための保存 ##
		Push, BasePointer,
		Mov, BasePointer, StackPointer,
		// ## 引数を含む変数領域の確保 ##
		Push, Integer(1), // iの1つ
		Pop, R1, // 一旦レジスタに入れてから引き算する
		Sub, StackPointer, R1, // SPをずらす
		// ## 引数と変数の結び付け ##
		// mainにはなし
		//
		// # i := 1をする #
		// < assign-beg >
		Push, Integer(1), // 右辺
		Pop, R1, // 右辺辺の取り出し
		Mov, StackRelativeOffset{BasePointer, -1}, R1,
		// < assign-end >
		//
		// < for-beg >
		// main_for_block_1_beg:
		DefLabel(8),
		// # if check_x(i, 15) {} をする #
		// < if-beg >
		// ## 左辺の計算をする ##
		// check(i, 15)
		Push, Integer(15), //arg2として
		Push, StackRelativeOffset{BasePointer, -1}, // arg1として
		// < call-beg >
		Call, Label(1), // check_x
		Push, Integer(2), // 引数分spを戻す
		Pop, R1,
		Add, StackPointer, R1, //
		Push, ACM1, // 左辺, 結果をスタックに入れてあげる
		// < call-end >
		// ## 右辺の計算をする ##
		Push, True,
		// ## check_x(i, 15) == true ##
		// < eq-beg >
		Pop, R2, // 右辺の取り出し
		Pop, R1, // 左辺の取り出し
		Eq, R1, R2, // r1 == r2
		Push, ZeroFlag,
		// < eq-end >
		// ## 条件分岐 ##
		Pop, Temporal1,
		Je, Label(9), // main_if_1_if, check_x(i, 15) == trueならifブロックへ
		Jmp, Label(10), // main_if_1_else, check_x(i, 15) == trueでないならelseブロックへ
		// main_if_1_if:
		DefLabel(9),
		Syscall, Write, StdOut, StackRelativeOffset{BasePointer, -1}, // print i
		Syscall, Write, StdOut, Character(' '),
		Syscall, Write, StdOut, Character('f'),
		Syscall, Write, StdOut, Character('i'),
		Syscall, Write, StdOut, Character('z'),
		Syscall, Write, StdOut, Character('z'),
		Syscall, Write, StdOut, Character('b'),
		Syscall, Write, StdOut, Character('u'),
		Syscall, Write, StdOut, Character('z'),
		Syscall, Write, StdOut, Character('z'),
		Syscall, Write, StdOut, Character('\n'),
		Jmp, Label(11), // main_if_1_end, ifの終了へ
		// main_if_1_else:
		DefLabel(10),
		// # if check(i, 5) {} をする #
		// < if-beg >
		// ## 左辺の計算をする ##
		// ### check_x(i, 5)をする ###
		Push, Integer(5), // 5をarg2として
		Push, StackRelativeOffset{BasePointer, -1}, // iをarg1として
		// < call-beg >
		Call, Label(1), // check_x
		Push, Integer(2), // 引数分spを戻してあげる
		Pop, R1,
		Add, StackPointer, R1,
		Push, ACM1, // 左辺, 結果をスタックに
		// < call-end >
		// ## 右辺の計算をする ##
		Push, True,
		// ## check(i, 5) == trueをする ##
		// < eq-beg >
		Pop, R2, // 右辺の取り出し
		Pop, R1, // 左辺の取り出し
		Eq, R1, R2, // r1 == r2
		Push, ZeroFlag, // 結果
		// < eq-end >
		// ## 条件分岐 ##
		Je, Label(101), // main_if_2_if, check(i, 5) == trueならifブロックへ
		Jmp, Label(102), // main_if_2_else, check(i, 5) == trueでないならelseブロックへ
		// main_if_2_if:
		DefLabel(101),
		Syscall, Write, StdOut, StackRelativeOffset{BasePointer, -1}, // print i
		Syscall, Write, StdOut, Character(' '),
		Syscall, Write, StdOut, Character('b'),
		Syscall, Write, StdOut, Character('u'),
		Syscall, Write, StdOut, Character('z'),
		Syscall, Write, StdOut, Character('z'),
		Syscall, Write, StdOut, Character('\n'),
		Jmp, Label(103), // main_if_2_end, ifの終了へ
		// main_if_2_else:
		DefLabel(102),
		// # if check(i, 3) {} をする #
		// < if-beg >
		// ## 左辺の計算をする ##
		// ### check_x(i, 3)をする ###
		Push, Integer(3), // 3をarg2として
		Push, StackRelativeOffset{BasePointer, -1}, // iをarg1として
		// < call-beg >
		Call, Label(1), // check_x
		Push, Integer(2), // 引数分spを戻す
		Pop, R1,
		Add, StackPointer, R1,
		Push, ACM1, // 左辺
		// < call-end >
		// ## 右辺の計算をする ##
		Push, True,
		// ## check(i, 3) == trueをする ##
		// < eq-beg >
		Pop, R2, // 右辺の取り出し
		Pop, R1, // 左辺の取り出し
		Eq, R1, R2, // r1 == r2
		Push, ZeroFlag, // 結果
		// < eq-end >
		Pop, Temporal1, // いらないので結果は捨てる
		Je, Label(201), // main_if_3_if, check(i, 3) == trueならifブロックへ
		Jmp, Label(202), // main_if_3_else, check(i, 3) == trueでないならelseブロックへ
		// main_if_3_if:
		DefLabel(201),
		Syscall, Write, StdOut, StackRelativeOffset{BasePointer, -1}, // print i
		Syscall, Write, StdOut, Character(' '),
		Syscall, Write, StdOut, Character('f'),
		Syscall, Write, StdOut, Character('i'),
		Syscall, Write, StdOut, Character('z'),
		Syscall, Write, StdOut, Character('z'),
		Syscall, Write, StdOut, Character('\n'),
		Jmp, Label(203), // main_if_3_end, ifの終了へ
		// main_if_3_else:
		DefLabel(202),
		Syscall, Write, StdOut, StackRelativeOffset{BasePointer, -1}, // print i
		Syscall, Write, StdOut, Character(' '),
		Syscall, Write, StdOut, Character('\n'),
		Jmp, Label(203), // main_if_3_end, ifの終了へ
		// main_if_3_end:
		DefLabel(203),
		// < if-end >
		//
		Jmp, Label(103), // main_if_2_end, ifの終了へ
		// main_if_2_end:
		DefLabel(103),
		// < if-end >
		Jmp, Label(11), // main_if_1_end, ifの終了へ
		// main_if_1_end:
		DefLabel(11),
		// < if-end >
		//
		// # i++をする #
		// < mov-beg >
		// ## 右辺を計算する ##
		// ### i + 1 ###
		Push, StackRelativeOffset{BasePointer, -1}, // 左辺
		Push, Integer(1), // 右辺
		// < add-beg >
		Pop, R2, // 右辺の取り出し
		Pop, R1, // 左辺の取り出し
		Add, R1, R2, // r1 += r2
		Push, R1, // 結果
		// < add-end >
		Pop, R1, // 結果を取り出す
		Mov, StackRelativeOffset{BasePointer, -1}, R1, // i = (i+1)
		// < mov-end >
		//
		// # if i == 101 {} をする #
		// < if-beg >
		// ## 左辺の計算をする ##
		Push, StackRelativeOffset{BasePointer, -1}, // 左辺i
		// ## 右辺の計算をする ##
		Push, Integer(101), // 右辺
		// < eq-beg >
		Pop, R2, // 右辺の取り出し
		Pop, R1, // 左辺の取り出し
		Eq, R1, R2, // r1 == r2
		Push, ZeroFlag,
		// < eq-end >
		// ## 条件分岐 ##
		Pop, Temporal1, // eqの結果はzero flagそのものを使うのでいらないので捨てる
		Je, Label(12), // main_if_4_if, (i==101) == trueならifブロックへ
		Jmp, Label(13), // main_if_4_else, (i==101) == falseならelseブロックへ
		// main_if_4_if:
		DefLabel(12),
		Jmp, Label(15), // main_for_block_1_end, break, forの終了へ
		Jmp, Label(14), // main_if_4_end, ifの終了へ
		// main_if_4_else:
		DefLabel(13),
		Jmp, Label(14), // main_if_4_end, ifの終了へ
		// main_if_4_end:
		DefLabel(14),
		Jmp, Label(8), // main_for_block_1_beg, forの最初へ
		// < if-end >
		//
		// main_for_block_1_end:
		DefLabel(15),
		// < for-end >
		//
		// # 戻り値の返却 #
		Push, Integer(0),
		Pop, R1,
		Mov, ACM1, R1,
		// # 関数の終了処理 #
		Mov, StackPointer, BasePointer,
		Pop, BasePointer,
		Ret,
	})
	assert.Nil(t, rt.CollectLabels())
	assert.Nil(t, rt.Run())

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	s := strings.TrimRight(buf.String(), "") // バッファーから文字列へ変換
	os.Stdout = tmpStdout

	fizzbuzz := `1 
2 
3 fizz
4 
5 buzz
6 fizz
7 
8 
9 fizz
10 buzz
11 
12 fizz
13 
14 
15 fizzbuzz
16 
17 
18 fizz
19 
20 buzz
21 fizz
22 
23 
24 fizz
25 buzz
26 
27 fizz
28 
29 
30 fizzbuzz
31 
32 
33 fizz
34 
35 buzz
36 fizz
37 
38 
39 fizz
40 buzz
41 
42 fizz
43 
44 
45 fizzbuzz
46 
47 
48 fizz
49 
50 buzz
51 fizz
52 
53 
54 fizz
55 buzz
56 
57 fizz
58 
59 
60 fizzbuzz
61 
62 
63 fizz
64 
65 buzz
66 fizz
67 
68 
69 fizz
70 buzz
71 
72 fizz
73 
74 
75 fizzbuzz
76 
77 
78 fizz
79 
80 buzz
81 fizz
82 
83 
84 fizz
85 buzz
86 
87 fizz
88 
89 
90 fizzbuzz
91 
92 
93 fizz
94 
95 buzz
96 fizz
97 
98 
99 fizz
100 buzz
`
	assert.Equal(t, fizzbuzz, s)
	assert.Equal(t, Integer(0), rt.reg[ACM1])
}

func TestRuntime_Run_Fibonacci(t *testing.T) {
	rt := NewRuntime(100, 10)
	rt.Load(Program{
		// func fib(n int) int {
		//   if n < 2 {
		//     return n
		//   }
		//   return fib(n-1) + fib(n-2)
		// }
		// func main() int {
		//   return fib(10)
		// }

		// fib(1):
		//   push bp		// 復元のために保存
		//   mov bp sp		// 引数を取り出すためにスタックの位置をずらす
		//   push 1			// 使用される変数の数だけspを下げる
		//   pop r1			//
		//   sub sp r1		// sp -= r1
		//   mov [bp-1] [bp+2]	// 引数と変数を結びつける, n:[bp-1] = arg1:[bp+2]
		//   // if n < 2
		//   < lt >
		//   push [bp-1]	// 左辺, n
		//   push 2			// 右辺
		//   pop r2			// 右辺の取り出し
		//   pop r1			// 左辺の取り出し
		//   lt r1 r2
		//   je if_if_block_jGcBdPTUNWrbPQrSxTuS(2) // zf==trueならif blockへ
		//   jmp if_end_gNEXgcrQiVuTXurmOGFW(3) // zf==trueでないならreturn blockへ
		// [ if block ]
		// if_if_block_jGcBdPTUNWrbPQrSxTuS(2):
		//   push [bp-1]	// nをこれまでの計算結果として
		//   < return >
		//   pop r1			// 計算結果をなぜかわからないけど一旦r1に入れる
		//   mov r10 r1		// 本来の戻り値レジスタに入れる
		//   mov sp bp		// spの復元
		//   pop bp			// bpの復元
		//   ret
		// if_end_gNEXgcrQiVuTXurmOGFW(3):
		//   < add >
		//   << sub >>		// addの左辺
		//   push [bp-1]	// 左辺, n
		//   push 1			// 右辺
		//   pop r2			// 右辺の取り出し
		//   pop r1			// 左辺の取り出し
		//   sub r1 r2		// n -= 1
		//   push r1		// 計算結果のプッシュ, 次のcallの引数
		//
		//   << call >>
		//   call fib(1)	// fib(n-1)
		//   push 1			// 呼び出し後の後処理, 1は引数の合計
		//   pop r1			//
		//   add sp r1		//
		//	 push r10		// fib(n-1)の戻り値をaddの左辺とする
		//
		//   << sub >>		// addの右辺
		//   push [bp-1]	// 左辺, n
		//   push 2			// 右辺
		//   pop r2			// 右辺の取り出し
		//   pop r1			// 左辺の取り出し
		//   sub r1 r2		// n -= 2
		//   push r1		// 計算結果のプッシュ, 次のcallの引数
		//
		//   << call >>
		//   call fib(1)	// fib(n-2)
		//   push 1			// 呼び出し後の後処理, 1は引数の合計
		//   pop r1			//
		//   add sp r1		//
		//   push r10		// fib(n-2)の戻り値をaddの右辺とする
		//
		//   pop r2			// 直前の結果を右辺
		//   pop r1			// 一個前の結果を左辺
		//   add r1 r2		// fib(n-1) += fib(n-2)
		//   push r1		// 結果をスタックへ
		//   < return >
		//   pop r1			// 計算結果をなぜかわからないけど一旦r1に入れる
		//   mov r10 r1		// 本来の戻り値レジスタに入れる
		//   mov sp bp		// spの復元
		//   pop bp			// bpの復元
		//   ret
		DefLabel(1),
		// 関数終了時の戻り場所を記録
		Push, BasePointer,
		Mov, BasePointer, StackPointer,
		// 関数内で使用される変数の数だけSPを下げる(変数領域の確保)
		Sub, StackPointer, Integer(1),
		// 引数と変数を結びつける(代入によって)
		Mov, StackRelativeOffset{BasePointer, -1}, StackRelativeOffset{BasePointer, +2},
		//  lt
		Push, StackRelativeOffset{BasePointer, -1},
		Push, Integer(2),
		Pop, R2,
		Pop, R1,
		Lt, R1, R2,
		Je, Label(2),
		Jmp, Label(3),
		//
		DefLabel(2),
		Push, StackRelativeOffset{BasePointer, -1},
		Pop, R1,
		Mov, R10, R1,
		Mov, StackPointer, BasePointer,
		Pop, BasePointer,
		Ret,
		//
		DefLabel(3),
		Push, StackRelativeOffset{BasePointer, -1},
		Push, Integer(1),
		Pop, R2,
		Pop, R1,
		Sub, R1, R2, //
		Push, R1,
		//
		Call, Label(1),
		Push, Integer(1),
		Pop, R1,
		Add, StackPointer, R1,
		Push, R10,
		//
		Pop, R2,
		Pop, R1,
		Add, R1, R2,
		Push, R1,
		//
		Pop, R1,
		Mov, R10, R1,
		Mov, StackPointer, BasePointer,
		Pop, BasePointer,
		Ret,

		// main(0):
		//   // 関数終了時の戻り場所を記録
		//   push bp		// 復元のために保存
		//   mov bp sp		// 引数を取り出すためにスタックの位置をずらす
		//   sub sp 0		// 引数分spをずらす, メインの引数は0
		//   // 関数内で使用される変数の数だけSPを下げる(変数領域の確保)
		//   push 0			//
		//   pop r1			//
		//   sub sp r1		//
		//
		//   push 10		// 呼び出し用の引数
		//   < call >
		//   call fib(1)
		//   push 1			// 呼び出し後の後処理, 1は引数の合計
		//   pop r1			//
		//   add sp r1		//
		//	 push r10		// fib(10)の戻り値を結果として
		//   < return >
		//   pop r1			// 計算結果をなぜかわからないけど一旦r1に入れる
		//   mov r10 r1		// 本来の戻り値レジスタに入れる
		//   mov sp bp		// spの復元
		//   pop bp			// bpの復元
		//   ret
		DefLabel(0),
		Push, BasePointer,
		Mov, BasePointer, StackPointer,
		Sub, StackPointer, Integer(0),
		//
		Push, Integer(0),
		Pop, R1,
		Sub, StackPointer, R1,
		//
		Push, Integer(10),
		//
		Call, Label(1),
		Push, Integer(1),
		Pop, R1,
		Add, StackPointer, R1,
		Push, R10,
		//
		Pop, R1,
		Mov, R10, R1,
		Mov, StackPointer, BasePointer,
		Pop, BasePointer,
		Ret,
	})
	assert.Nil(t, rt.CollectLabels())
	assert.Nil(t, rt.Run())
	assert.Equal(t, Integer(55), rt.reg[R10])
}

func TestRuntime_Run_var(t *testing.T) {
	rt := NewRuntime(100, 10)
	rt.Load(Program{
		DefLabel(0),
		Push, BasePointer, // 復元のため
		Mov, BasePointer, StackPointer,
		// ## 引数を含む変数領域の確保の用意(おそらく) ##
		Push, Integer(0), // 関数ないの変数n,x用の領域の確保
		Pop, R1,
		Sub, StackPointer, R1, //

		Mov, ACM1, Integer(0),

		// return
		Mov, StackPointer, BasePointer, // spの復元
		Pop, BasePointer, // bpの復元
		Ret,
	})
	assert.Nil(t, rt.CollectLabels())
	assert.Nil(t, rt.Run())
	assert.Equal(t, Integer(0), rt.reg[ACM1])
}
