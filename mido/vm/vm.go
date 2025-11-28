package vm

import (
	"errors"
	"fmt"
	"io"
	"runtime/debug"
	"strings"
	"sudonters/libzootr/mido/code"
	"sudonters/libzootr/mido/compiler"
	"sudonters/libzootr/mido/objects"

	"github.com/etc-sudonters/substrate/dontio"
)

type VM struct {
	Objects *objects.Table
	Funcs   objects.BuiltInFunctions
	Std     *dontio.Std
	ChkQty  objects.BuiltInFunction
}

func (this *VM) Execute(bytecode compiler.Bytecode) (obj objects.Object, err error) {
	result := objects.Null
	unit := execution{0, bytecode, newstack[objects.Object](256)}
	EOT := unit.endOfTape()

	defer func() {
		if r := recover(); r != nil {
			if thisErr, ok := r.(error); ok {
				err = thisErr
			} else if str, ok := r.(string); ok {
				err = fmt.Errorf(str)
			}

			this.Std.WriteLineErr("VM panicked: %s\n%s", err, debug.Stack())
		}
	}()

loop:
	for unit.ip < EOT {
		thisOp := unit.readOp()
		switch thisOp {
		case code.NOP:
			continue
		case code.ERR:
			err = errors.Join(errors.New("execution halted"), err)
			break loop
		case code.PUSH_T:
			unit.stack.push(objects.PackedTrue)
		case code.PUSH_F:
			unit.stack.push(objects.PackedFalse)
		case code.PUSH_CONST, code.PUSH_FUNC, code.PUSH_PTR, code.PUSH_STR:
			index := unit.readIndex()
			unit.stack.push(this.Objects.AtIndex(index))
		case code.INVERT:
			obj := unit.stack.pop()
			unit.stack.push(objects.PackBool((!this.Truthy(obj))))
		case code.NEED_ALL:
			count := int(unit.readu16())
			reduction := true
			stackargs := unit.stackargs(count)
			for _, obj := range stackargs {
				if !this.Truthy(obj) {
					reduction = false
					break
				}
			}
			unit.stack.popN(count)
			unit.stack.push(objects.PackBool(reduction))
		case code.NEED_ANY:
			count := int(unit.readu16())
			reduction := false
			stackargs := unit.stackargs(count)
			for _, obj := range stackargs {
				if this.Truthy(obj) {
					reduction = true
					break
				}
			}
			unit.stack.popN(count)
			unit.stack.push(objects.PackBool(reduction))
		case code.INVOKE:
			answer := objects.Null
			obj := unit.stack.pop()
			count := int(unit.readu16())
			args := unit.stackargs(count)
			answer, err = this.Funcs.Call(this.Objects, obj, args)
			unit.stack.popN(count)
			if err != nil {
				break loop
			}
			if answer != objects.Null {
				unit.stack.push(answer)
			}
		case code.INVOKE_0:
			index := unit.readIndex()
			obj := this.Objects.AtIndex(index)
			answer, err := this.Funcs.Call(this.Objects, obj, nil)
			if err != nil {
				break loop
			}
			if answer != objects.Null {
				unit.stack.push(answer)
			}
		case code.CHK_QTY:
			if this.ChkQty == nil {
				err = fmt.Errorf("fastop 0x%02X not found in table", thisOp)
				break loop
			}

			index := unit.readIndex()
			qty := unit.readu8()
			obj := this.Objects.AtIndex(index)
			answer, err := this.ChkQty(this.Objects, []objects.Object{
				obj, objects.PackF64(float64(qty)),
			})
			if err != nil {
				break loop
			}
			unit.stack.push(answer)
		case code.CMP_EQ, code.CMP_NQ, code.CMP_LT:
			err = fmt.Errorf("runtime comparison not implemented")
			break loop
		default:
			err = fmt.Errorf("unrecognized op: 0x%02x", thisOp)
			break loop
		}
	}

	if err == nil && unit.stack.ptr > 0 {
		result = unit.stack.pop()
	}
	if err != nil {
		this.Std.WriteLineErr("VM Error: %s", err)
	}
	return result, err
}

const warning dontio.ForegroundColor = 9

func (this *VM) Truthy(obj objects.Object) bool {
	if obj != objects.PackedTrue && obj != objects.PackedFalse {
		fmt.Fprintf(this.Std.Err, warning.Paint("truthy checked non-boolean %q %X\n"), obj.Type(), obj)
	}

	return obj.Truthy()
}

type Disassembly struct {
	Constants []Constant
	Code      code.Instructions
	Dis       string
}

func Disassemble(bytecode compiler.Bytecode, objs *objects.Table) Disassembly {
	var dis Disassembly
	dis.Dis = code.DisassembleToString(bytecode.Tape)
	dis.Code = bytecode.Tape
	dis.Constants = make([]Constant, len(bytecode.Consts))
	for i := range dis.Constants {
		c := Constant{
			Index:  bytecode.Consts[i],
			Object: objs.AtIndex(bytecode.Consts[i]),
		}

		switch c.Object.Type() {
		case objects.STR_PTR32:
			c.Value = objects.UnpackPtr32(c.Object)
			c.Name = bytecode.Names[c.Index]
			break
		case objects.STR_STR32:
			c.Value = objs.DerefString(c.Object)
			break
		case objects.STR_BYTES:
			c.Value = objects.UnpackBytes(c.Object)
			break
		case objects.STR_BOOL:
			c.Value = objects.UnpackBool(c.Object)
			break
		case objects.STR_F64:
			c.Value = objects.UnpackF64(c.Object)
			break

		}

		dis.Constants[i] = c
	}

	return dis
}

type Constant struct {
	Index  objects.Index
	Object objects.Object
	Name   string
	Value  any
}

func (this Constant) String() string {
	var view strings.Builder
	obj := this.Object
	constant := this.Index
	fmt.Fprintf(&view, "0x%04X:\t0x%08X\n", constant, obj)
	ty := obj.Type()
	fmt.Fprintf(&view, "\ttype:\t%s\n", ty)
	switch ty {
	case objects.STR_PTR32:
		ptr := this.Value.(objects.Ptr32)
		fmt.Fprintf(&view, "\tname:\t%q\n", this.Name)
		fmt.Fprintf(&view, "\ttag:\t%s\tptr:\t%04X\n", ptr.Tag, ptr.Addr)
		break
	case objects.STR_STR32:
		fmt.Fprintf(&view, "\tvalue:\t%q\n", this.Value.(string))
		break
	case objects.STR_BYTES:
		fmt.Fprintf(&view, "\tvalue:\t%v\n", this.Value.(objects.Bytes))
		break
	case objects.STR_BOOL:
		fmt.Fprintf(&view, "\tvalue:\t%t\n", this.Value.(bool))
		break
	case objects.STR_F64:
		fmt.Fprintf(&view, "\tvalue:\t%f\n", this.Value.(float64))
		break
	}

	return view.String()
}

func (this *VM) Dis(w io.Writer, bytecode compiler.Bytecode) {
	code.DisassembleInto(w, bytecode.Tape)
	if len(bytecode.Consts) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "CONSTANTS")
		for _, constant := range bytecode.Consts {
			obj := this.Objects.AtIndex(constant)
			fmt.Fprintf(w, "0x%04X:\t0x%08X\n", constant, obj)
			ty := obj.Type()
			fmt.Fprintf(w, "\ttype:\t%s\n", ty)
			switch ty {
			case objects.STR_PTR32:
				ptr := objects.UnpackPtr32(obj)
				name := bytecode.Names[constant]
				fmt.Fprintf(w, "\tname:\t%q\n", name)
				fmt.Fprintf(w, "\ttag:\t%s\n\tptr:\t%04X\n", ptr.Tag, ptr.Addr)
				break
			case objects.STR_STR32:
				fmt.Fprintf(w, "\tvalue:\t%q\n", this.Objects.DerefString(obj))
				break
			case objects.STR_BYTES:
				fmt.Fprintf(w, "\tvalue:\t%v\n", objects.UnpackBytes(obj))
				break
			case objects.STR_BOOL:
				fmt.Fprintf(w, "\tvalue:\t%t\n", objects.UnpackBool(obj))
				break
			case objects.STR_F64:
				fmt.Fprintf(w, "\tvalue:\t%f\n", objects.UnpackF64(obj))
				break
			}
			fmt.Fprintln(w)
		}
	}
}
