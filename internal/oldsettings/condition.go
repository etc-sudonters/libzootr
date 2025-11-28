package oldsettings

import (
	"errors"
	"fmt"
)

// this is actually two uint8
// the upper bits describe what the condition is
// the lower bits describe how much of that condition
type quantitycondition struct {
	q uint16
}

type qc interface {
	LacsCondition | BridgeCondition | GanonBKCondition
}

func (q quantitycondition) Decode() (which Condition, qty uint8) {
	which = Condition((q.q & 0xFF00) >> 8)
	qty = uint8(0x00FF & q.q)
	return
}

func encodeqty(which Condition, qty uint8) quantitycondition {
	return quantitycondition{(uint16(which) << 8) | uint16(qty)}
}

func DecodeCondition[C qc](c C) (which Condition, qty uint8) {
	return quantitycondition(c).Decode()
}

func ExpectedCondition[C qc](c C, which Condition) (uint8, bool) {
	cond, qty := DecodeCondition(c)
	if cond != which {
		return 0, false
	}

	return qty, true
}

type LacsCondition quantitycondition
type BridgeCondition quantitycondition
type GanonBKCondition quantitycondition

func CreateLacs(cond Condition, qty uint8) LacsCondition {
	return LacsCondition(encodeqty(cond, qty))
}

func CreateBridge(cond Condition, qty uint8) BridgeCondition {
	return BridgeCondition(encodeqty(cond, qty))
}

func CreateGanonBK(cond Condition, qty uint8) GanonBKCondition {
	return GanonBKCondition(encodeqty(cond, qty))
}

type Condition uint8

const (
	CondUnitialized Condition = iota
	CondDefault
	CondMedallions
	CondStones
	CondRewards
	CondTokens
	CondHearts
	CondOpen
	CondVanilla
	CondTriforce
)

func (this Condition) String() string {
	switch this {
	case CondUnitialized:
		panic(errors.New("uninitialized condition"))
	case CondMedallions:
		return "medallions"
	case CondStones:
		return "stones"
	case CondRewards:
		return "rewards"
	case CondTokens:
		return "tokens"
	case CondHearts:
		return "hearts"
	case CondDefault:
		return "default"
	case CondOpen:
		return "open"
	case CondVanilla:
		return "vanilla"
	case CondTriforce:
		return "triforce"
	default:
		panic(fmt.Errorf("unknown condition flag %x", uint8(this)))
	}
}
