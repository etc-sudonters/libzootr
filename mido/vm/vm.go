package vm

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime/debug"
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
				err = errors.New(str)
			}

			err = fmt.Errorf("PANIC!!!! %w\n%s", err, debug.Stack())
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
	return result, err
}

func (this *VM) Truthy(obj objects.Object) bool {
	if obj != objects.PackedTrue && obj != objects.PackedFalse {
		slog.Warn("truthy checked non-boolean", "kind", obj.Type(), "obj", obj.String())
	}

	return obj.Truthy()
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
