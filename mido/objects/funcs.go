package objects

type BuiltInFunctionDef struct {
	Name       string
	ParamCount int
}
type BuiltInFunction func(*Table, []Object) (Object, error)
type BuiltInFunctions []BuiltInFunction

func (this BuiltInFunctions) Call(tbl *Table, callee Object, args []Object) (Object, error) {
	ptr := UnpackPtr32(callee)
	return this[ptr.Addr](tbl, args)
}
