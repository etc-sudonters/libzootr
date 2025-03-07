package script

import "fmt"

func Unify(t1, t2 Type, subs Substitutions) error {
	if t1.StrictlyEq(t2) {
		return nil
	}

	if tv, isTv := t1.(TypeVar); isTv {
		return unifyVar(tv, t2, subs)
	}

	if tv, isTv := t2.(TypeVar); isTv {
		return unifyVar(tv, t1, subs)
	}

	tt1, isTT1 := t1.(TypeTuple)
	tt2, isTT2 := t2.(TypeTuple)
	if isTT1 == isTT2 {
		return unifyTuples(tt1, tt2, subs)
	}

	return CannotUnify{t1, t2}
}

func Annotate(node AstNode, env *AstEnv, subs Substitutions) error {
	switch node := node.(type) {
	case *FindNode:
		if err := AnnotateAll(node.Finding, env, subs); err != nil {
			return fmt.Errorf("on finding select clause: %w", err)
		}
		if err := AnnotateAll(node.Clauses, env, subs); err != nil {
			return fmt.Errorf("on finding predicate clause: %w", err)
		}
		if err := AnnotateAll(node.Rules, env, subs); err != nil {
			return fmt.Errorf("on finding rule clause: %w", err)
		}

		tv := NextTypeVar()
		tt := TypeTuple{GetTypes(node.Finding)}
		node.Type = tv
		subs[tv] = tt
		if err := Unify(tv, tt, subs); err != nil {
			return fmt.Errorf("while unifying finding: %w", err)
		}
	case *InsertNode:
		if err := AnnotateAll(node.Inserting, env, subs); err != nil {
			return fmt.Errorf("on insert select clause: %w", err)
		}
		if err := AnnotateAll(node.Clauses, env, subs); err != nil {
			return fmt.Errorf("on insert predicate clause: %w", err)
		}
		if err := AnnotateAll(node.Rules, env, subs); err != nil {
			return fmt.Errorf("on insert rule clause: %w", err)
		}

		node.Type = TypeVoid{}
	case *RuleDeclNode:
		newEnv := NewAstEnv()
		newEnv.parent = env
		if err := AnnotateAll(node.Params, newEnv, subs); err != nil {
			return fmt.Errorf("on clause param: %w", err)
		}
		if err := AnnotateAll(node.Clauses, newEnv, subs); err != nil {
			return fmt.Errorf("on clause predicate: %w", err)
		}

		ty := env.GetNamed(node.Name)
		if ty == nil {
			tv := NextTypeVar()
			ty = tv
			subs[tv] = TypeTuple{GetTypes(node.Params)}
			env.AddNamed(node.Name, ty)
		}
		node.Type = ty
		node.Env = *newEnv
	case *InsertTripletNode:
		return AnnotateTriplet(&node.TripletNode, env, subs)
	case *TripletClauseNode:
		return AnnotateTriplet(&node.TripletNode, env, subs)
	case *RuleClauseNode:
		if err := AnnotateAll(node.Args, env, subs); err != nil {
			return fmt.Errorf("on clause args: %w", err)
		}
		args := TypeTuple{GetTypes(node.Args)}
		ty := env.GetNamed(node.Name)
		if ty == nil {
			tv := NextTypeVar()
			ty = tv
			subs[tv] = args
			env.AddNamed(node.Name, ty)
		}
		node.Type = ty
		if err := Unify(ty, args, subs); err != nil {
			return fmt.Errorf("when unifying rule: %w", err)
		}
	case *VarNode:
		ty := env.GetNamed(node.Name)
		if ty == nil {
			ty = NextTypeVar()
			env.AddNamed(node.Name, ty)
		}
		node.Type = ty
	case *EntityNode:
		node.Type = TypeNumber{}
		if node.Var != nil {
			ty := env.GetNamed(node.Var.Name)
			if ty == nil {
				ty = NextTypeVar()
				env.AddNamed(node.Var.Name, ty)
			}
			if tv, isTv := ty.(TypeVar); isTv {
				subs[tv] = TypeNumber{}
			}
			node.Var.Type = ty
		}
	case *AttrNode:
		ty, id := env.GetAttr(node.Name)
		node.Id = id
		node.Type = ty
		if node.Type == nil || node.Id == 0 {
			return fmt.Errorf("unknown attribute %q", node.Name)
		}
	case *BoolNode, *NumberNode, *StringNode:
	default:
		return fmt.Errorf("unknown node type: %#v", node)
	}
	return nil
}

func AnnotateAll[T AstNode](nodes []T, env *AstEnv, subs Substitutions) error {
	for i := range nodes {
		if err := Annotate(nodes[i], env, subs); err != nil {
			return err
		}
	}
	return nil
}

func AnnotateTriplet(node *TripletNode, env *AstEnv, subs Substitutions) error {
	/* TODO triplets have the type [ Number Var{X} Var{X} ] Var{Y}
	   All instances of [ _ ATTR_NAME _ ] need to unify under this pattern
	*/

	if err := Annotate(node.Id, env, subs); err != nil {
		return fmt.Errorf("on triplet id: %w", err)
	}
	if err := Annotate(node.Attribute, env, subs); err != nil {
		return fmt.Errorf("on triplet attribute: %w", err)
	}
	if err := Annotate(node.Predicate, env, subs); err != nil {
		return fmt.Errorf("on triplet predicate: %w", err)
	}
	attrType := node.Attribute.GetType()
	tripTT, err := Substitute(TypeTuple{[]Type{node.Id.GetType(), node.Attribute.GetType(), node.Predicate.GetType()}}, subs)
	if err != nil {
		return err
	}
	attrTT := TypeTuple{[]Type{TypeNumber{}, attrType, attrType}}
	if err != nil {
		return err
	}

	if err := Unify(tripTT, attrTT, subs); err != nil {
		return err
	}
	tv := NextTypeVar()
	subs[tv] = attrTT
	node.Type = tv
	return Unify(node.Type, attrTT, subs)
}

type CannotUnify struct {
	T1, T2 Type
}

func (this CannotUnify) Error() string {
	return fmt.Sprintf("cannot unify %s and %s", this.T1, this.T2)
}

func unifyTuples(tt1, tt2 TypeTuple, subs Substitutions) error {
	if len(tt1.Types) != len(tt2.Types) {
		return CannotUnify{tt1, tt2}
	}

	var err error
	for i := range tt1.Types {
		if err = Unify(tt1.Types[i], tt2.Types[i], subs); err != nil {
			return err
		}
	}

	return nil
}

func unifyVar(tv TypeVar, ty Type, subs Substitutions) error {
	if ty2, exists := subs[tv]; exists {
		return Unify(ty2, ty, subs)
	}
	if tv2, isTv := ty.(TypeVar); isTv {
		if ty, exists := subs[tv2]; exists {
			return Unify(tv, ty, subs)
		}
	}

	if reoccurs(tv, ty, subs) {
		return TypeReoccurs{tv}
	}

	subs[tv] = ty
	return nil
}

func reoccurs(tv TypeVar, ty Type, subs Substitutions) bool {
	if tv.StrictlyEq(ty) {
		return true
	}

	if tv2, isTv := ty.(TypeVar); isTv {
		if ty, exists := subs[tv2]; exists {
			return reoccurs(tv, ty, subs)
		}
	}

	if tt, isTT := ty.(TypeTuple); isTT {
		for i := range tt.Types {
			if reoccurs(tv, tt.Types[i], subs) {
				return true
			}
		}
	}

	return false
}

type TypeReoccurs struct {
	Var TypeVar
}

func (this TypeReoccurs) Error() string {
	return fmt.Sprintf("%s reoccurs in itself", this.Var)
}
