package main

import (
    `fmt`
    `strconv`
    `strings`
    `unsafe`
)

type (
	OpCode   uint8
    LetKind  uint8
    Program  []Instr
    Compiler struct{}
)

const (
    Lambda = "λ"
)

const (
    _ OpCode = iota
    OP_ldconst          // ldconst      <val>       : Push <val> onto stack.
    OP_ldvar            // ldvar        <name>      : Push the content of variable <name> onto stack.
    OP_define           // define       <name>      : Define a global variable <name> with content at the stack top.
    OP_set              // set          <name>      : Set the variable <name> to content at the stack top.
    OP_car              // car                      : Get the first part of a pair.
    OP_cdr              // cdr                      : Get the second part of a pair.
    OP_cons             // cons                     : Construct a new pair from stack top.
    OP_eval             // eval                     : Evaluate the value on stack top.
    OP_drop             // drop                     : Drop one value from stack.
    OP_goto             // goto         <pc>        : Goto <pc> unconditionally.
    OP_if_false         // if_false     <pc>        : If the stack top is #f, goto <pc>.
    OP_apply            // apply        <argc>      : Apply procedure on stack with <argc> arguments.
    OP_return           // return                   : Return from procedure.
)

const (
    Let LetKind = iota
    LetRec
    LetStar
)

type Instr struct {
    u0 uint32
    u1 uint32
    p0 unsafe.Pointer
    p1 unsafe.Pointer
}

func (self Instr) Iv() uint32 { return self.u1 }
func (self Instr) Op() OpCode { return OpCode(self.u0) }
func (self Instr) Rv() Value  { return mkval(self.p0, self.p1).pack() }
func (self Instr) Sv() string { return mkstr(self.p0, int(self.u1)).String() }

func (self Instr) String() string {
    switch self.Op() {
        case OP_ldconst : return fmt.Sprintf("ldconst     %s", self.Rv())
        case OP_ldvar   : return fmt.Sprintf("ldvar       %s", self.Sv())
        case OP_define  : return fmt.Sprintf("define      %s", self.Sv())
        case OP_set     : return fmt.Sprintf("set         %s", self.Sv())
        case OP_car     : return "car"
        case OP_cdr     : return "cdr"
        case OP_cons    : return "cons"
        case OP_eval    : return "eval"
        case OP_drop    : return "drop"
        case OP_goto    : return fmt.Sprintf("goto       @%d", self.Iv())
        case OP_if_false: return fmt.Sprintf("if_false   @%d", self.Iv())
        case OP_apply   : return fmt.Sprintf("apply       %d", self.Iv())
        case OP_return  : return "return"
        default         : return fmt.Sprintf("OpCode(%d)", self.Op())
    }
}

func mku1(iv uint32, sv string) uint32 {
    if sv == "" {
        return iv
    } else if iv == 0 {
        return uint32(len(sv))
    } else {
        panic("fatal: encoding confliction between iv and sv")
    }
}

func mkp0(sv string, rv Value) unsafe.Pointer {
    if sv == "" && rv == nil {
        return nil
    } else if sv != "" && rv == nil {
        return straddr(sv)
    } else if sv == "" && rv != nil {
        return valitab(rv)
    } else {
        panic("fatal: encoding confliction between sv and rv")
    }
}

func mkins(op OpCode, iv uint32, sv string, rv Value) Instr {
    return Instr {
        u0: uint32(op),
        u1: mku1(iv, sv),
        p0: mkp0(sv, rv),
        p1: valaddr(rv),
    }
}

func (self *Program) add(op OpCode)             { *self = append(*self, mkins(op, 0, "", nil)) }
func (self *Program) val(op OpCode, val Value)  { *self = append(*self, mkins(op, 0, "", val)) }
func (self *Program) i32(op OpCode, val uint32) { *self = append(*self, mkins(op, val, "", nil)) }
func (self *Program) str(op OpCode, val string) { *self = append(*self, mkins(op, 0, val, nil)) }

func (self Program) String() string {
    var idx int
    var val Instr
    var buf []string

    /* empty program */
    if len(self) == 0 {
        return "(empty program)"
    }

    /* count the maximum digits */
    nb := len(self) - 1
    nd := len(strconv.Itoa(nb))
    fs := fmt.Sprintf("%%%dd :  %%s\n", nd)

    /* disassemble every instruction */
    for idx, val = range self {
        buf = append(buf, fmt.Sprintf(fs, idx, val))
    }

    /* pack into result */
    return strings.Join(buf, "")
}

func (self Compiler) Compile(src *List) (p Program) {
    self.compileList(&p, src)
    return
}

/** Sub-type Compiling **/

func (self Compiler) compileSet(p *Program, v *List) {
    var ok bool
    var sn Atom
    var vv *List

    /* unpack the variable name and value */
    if sn, ok = v.Car.(Atom) ; !ok { panic("compile: malformed let! construct: " + v.String()) }
    if vv, ok = v.Cdr.(*List); !ok { panic("compile: malformed let! construct: " + v.String()) }
    if vv.Cdr != nil               { panic("compile: malformed let! construct: " + v.String()) }

    /* emit the opcode */
    self.compileValue(p, vv.Car)
    p.val(OP_set, sn)
}

func (self Compiler) compileList(p *Program, v *List) {
    var ok bool
    var at Atom
    var vv *List

    /* empty list evaluates to nil */
    if v == nil {
        p.val(OP_ldconst, nil)
        return
    }

    /* (car v) is not an atom, apply the list immediately */
    if at, ok = v.Car.(Atom); !ok {
        p.i32(OP_apply, self.compileArgs(p, v, -1))
        return
    }

    /* must be a proper list to be applicable */
    if vv, ok = AsList(v.Cdr); !ok {
        panic("compile: improper list is not applicable: " + v.String())
    }

    /* check for built-in atoms */
    switch at {
        case "car"    : self.compileArgs(p, vv, 1); p.add(OP_car)
        case "cdr"    : self.compileArgs(p, vv, 1); p.add(OP_cdr)
        case "cons"   : self.compileArgs(p, vv, 2); p.add(OP_cons)
        case "eval"   : self.compileArgs(p, vv, 1); p.add(OP_eval)
        case "set!"   : self.compileSet(p, vv)
        case "begin"  : self.compileBlock(p, vv)
        case "quote"  : self.compileQuote(p, vv)
        case Lambda   : fallthrough
        case "lambda" : self.compileLambda(p, vv)
        case "define" : self.compileDefine(p, vv)
        case "if"     : self.compileCondition(p, vv)
        case "do"     : self.compileList(p, self.desugarDo(vv))
        case "let"    : self.compileList(p, self.desugarLet(vv, Let))
        case "let*"   : self.compileList(p, self.desugarLet(vv, LetStar))
        case "letrec" : self.compileList(p, self.desugarLet(vv, LetRec))
        default       : p.i32(OP_apply, self.compileArgs(p, v, -1))
    }
}

func (self Compiler) compileArgs(p *Program, v *List, n int) uint32 {
    var nb int
    var ok bool
    var vv *List

    /* scan every element */
    for s := v; s != nil; s, nb = vv, nb + 1 {
        if vv, ok = AsList(s.Cdr); !ok {
            panic("compile: improper list is not applicable: " + v.String())
        } else {
            self.compileValue(p, s.Car)
        }
    }

    /* apply the tuple */
    if n < 0 || n == nb {
        return uint32(nb)
    } else {
        panic(fmt.Sprintf("compile: expect %d arguments, got %d.", n, nb))
    }
}

func (self Compiler) compileBlock(p *Program, v *List) {
    for v != nil {
        ok := false
        self.compileValue(p, v.Car)

        /* drop the value if not the last one */
        if v.Cdr != nil {
            p.add(OP_drop)
        }

        /* check for proper list */
        if v, ok = AsList(v.Cdr); !ok {
            panic("compile: block must be a proper list: " + v.String())
        }
    }
}

func (self Compiler) compileValue(p *Program, v Value) {
    if v.IsIdentity() {
        p.val(OP_ldconst, v)
    } else if at, ok := v.(Atom); ok {
        p.val(OP_ldvar, at)
    } else if sl, ok := AsList(v); ok {
        self.compileList(p, sl)
    } else {
        panic("fatal: compile: invalid value type")
    }
}

func (self Compiler) compileQuote(p *Program, v *List) {
    if v != nil && v.Cdr == nil {
        p.val(OP_ldconst, v.Car)
    } else {
        panic("compile: `quote` takes exact 1 argument: " + v.String())
    }
}

func (self Compiler) compileLambda(p *Program, v *List) {

}

func (self Compiler) compileDefine(p *Program, v *List) {

}

func (self Compiler) compileCondition(p *Program, v *List) {

}

/** Syntax Desugaring **/

func (self Compiler) desugarDo(v *List) *List {
    var decl *List
    var cond *List
    var body *List
    var defs []Value
    var init []Value
    var step []Value

    /* list header */
    p := v
    ok := false

    /* deconstruct the list */
    if p == nil                      { panic("compile: malformed do construct: " + v.String()) }
    if decl, ok = p.Car.(*List); !ok { panic("compile: malformed do construct: " + v.String()) }
    if p   , ok = p.Cdr.(*List); !ok { panic("compile: malformed do construct: " + v.String()) }
    if cond, ok = p.Car.(*List); !ok { panic("compile: malformed do construct: " + v.String()) }
    if body, ok = p.Cdr.(*List); !ok { panic("compile: malformed do construct: " + v.String()) }

    /* parse the declarations */
    for p = decl; p != nil; {
        var s Atom
        var q *List
        var i Value
        var r Value

        /* get the initialization list, and move to next item */
        if q, ok = p.Car.(*List); !ok { panic("compile: malformed do construct: " + decl.String()) }
        if p, ok = AsList(p.Cdr); !ok { panic("compile: malformed do construct: " + decl.String()) }
        if s, ok = q.Car.(Atom) ; !ok { panic("compile: malformed do construct: " + decl.String()) }
        if q, ok = q.Cdr.(*List); !ok { panic("compile: malformed do construct: " + decl.String()) }

        /* check for the optional "step" part */
        if i, r = q.Car, s; q.Cdr != nil {
            if q, ok = q.Cdr.(*List); !ok { panic("compile: malformed do construct: " + decl.String()) }
            if r, ok = q.Car.(*List); !ok { panic("compile: malformed do construct: " + decl.String()) }
            if q.Cdr != nil               { panic("compile: malformed do construct: " + decl.String()) }
        }

        /* add to initialzer list */
        defs = append(defs, s)
        init = append(init, i)
        step = append(step, r)
    }

    /* check the condition expression */
    if p, ok = AsList(cond.Cdr); !ok { panic("compile: malformed do construct: " + cond.String()) }
    if p != nil && p.Cdr != nil      { panic("compile: malformed do construct: " + cond.String()) }

    /* rebuild the "do" construct */
    if p == nil {
        return self.rebuildDo(defs, init, step, cond.Car, nil, body)
    } else {
        return self.rebuildDo(defs, init, step, cond.Car, p.Car, body)
    }
}

func (self Compiler) desugarLet(v *List, kind LetKind) *List {
    var decl *List
    var body *List
    var defs []Value
    var init []Value

    /* list header */
    p := v
    n := 0
    ok := false

    /* deconstruct the list, body cannot be empty */
    if p == nil                      { panic("compile: malformed let construct: " + v.String()) }
    if decl, ok = AsList(p.Car); !ok { panic("compile: malformed let construct: " + v.String()) }
    if body, ok = p.Cdr.(*List); !ok { panic("compile: malformed let construct: " + v.String()) }

    /* parse the declarations */
    for p = decl; p != nil; n++ {
        var s Atom
        var q *List

        /* get the pair, and move to next item */
        if q, ok = p.Car.(*List); !ok { panic("compile: malformed let construct: " + decl.String()) }
        if p, ok = AsList(p.Cdr); !ok { panic("compile: malformed let construct: " + decl.String()) }
        if s, ok = q.Car.(Atom) ; !ok { panic("compile: malformed let construct: " + decl.String()) }
        if q, ok = q.Cdr.(*List); !ok { panic("compile: malformed let construct: " + decl.String()) }
        if q.Cdr != nil               { panic("compile: malformed let construct: " + decl.String()) }

        /* add to initializer list */
        defs = append(defs, s)
        init = append(init, q.Car)
    }

    /* need special handling of `letrec` */
    switch kind {
        case Let     : break
        case LetRec  : return self.rebuildLetRec(defs, init, body)
        case LetStar : n = 1
        default      : panic("fatal: invalid let kind")
    }

    /* rebuild as immediate lambda application expression */
    for i := len(defs) - n; i >= 0; i-- {
        arg := MakeList(defs[i:i + n]...)
        ref := MakePair(Atom(Lambda), MakePair(arg, body))
        body = MakeList(append([]Value{ref}, init[i:i + n]...)...)
    }

    /* all done */
    return body
}

/** Core Language Rebuilding **/

func (self Compiler) rebuildDo(defs []Value, init []Value, step []Value, cond Value, retv Value, body *List) *List {
    var pb, qb *List
    var pd, qd *List
    var pi, qi *List
    var ps, qs *List

    /* construct an unique name */
    ok := true
    name := fmt.Sprintf("#[desugar-do-%d]", nextid())

    /* loop variable, initial values and stepping */
    for _, v := range defs { AppendValue(&pd, &qd, v) }
    for _, v := range init { AppendValue(&pi, &qi, v) }
    for _, v := range step { AppendValue(&ps, &qs, v) }

    /* copy the body */
    for p := body; ok && p != nil; p, ok = AsList(p.Cdr) {
        AppendValue(&pb, &qb, p.Car)
    }

    /* check for loop body */
    if !ok || body == nil {
        panic("compile: loop body must be a proper list: " + body.String())
    }

    /* return an empty list if not specified */
    if retv == nil {
        retv = MakeList(Atom("quote"), nil)
    }

    /* append the loop recursion */
    AppendValue(&pb, &qb, MakePair(
        Atom(name),
        ps,
    ))

    /* recursive lambda body */
    loop := MakeList(
        Atom("if"),
        cond,
        retv,
        MakePair(Atom("begin"), pb),
    )

    /* reconstruct "do" with "lecrec" */
    return MakePair(Atom("letrec"), MakePair(
        MakeList(MakeList(Atom(name), MakeList(Atom("lambda"), pd, loop))),
        MakeList(MakePair(Atom(name), pi)),
    ))
}

func (self Compiler) rebuildLetRec(defs []Value, init []Value, body *List) *List {
    var pd, qd *List
    var pi, qi *List

    /* variable definations */
    for _, v := range defs {
        AppendValue(&pd, &qd, MakeList(v, MakeList(Atom("quote"), nil)))
    }

    /* set initial values */
    for i, v := range init {
        AppendValue(&pi, &qi, MakeList(Atom("set!"), defs[i], v))
    }

    /* reconstruct "letrec" with "let" and "set!" */
    qi.Cdr = body
    return MakePair(Atom("let"), MakePair(pd, pi))
}
