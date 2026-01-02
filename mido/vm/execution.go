package vm

import (
	"sudonters/libzootr/mido/code"
	"sudonters/libzootr/mido/compiler"
	"sudonters/libzootr/mido/objects"
)

type execution struct {
	ip    int
	code  compiler.Bytecode
	stack stack[objects.Object]
}

func (this *execution) readu8() uint8 {
	u8 := code.ReadU8(this.code.Tape[this.ip:])
	this.ip++
	return u8
}

func (this *execution) readu16() uint16 {
	u16 := code.ReadU16(this.code.Tape[this.ip:])
	this.ip += 2
	return u16
}

func (this *execution) endOfTape() int {
	return len(this.code.Tape)
}

func (this *execution) readIndex() objects.Index {
	return objects.Index(this.readu16())
}

func (this *execution) readOp() code.Op {
	return code.Op(this.readu8())
}

func (this *execution) stackargs(count int) []objects.Object {
	return this.stack.slice(this.stack.ptr-count, count)
}
