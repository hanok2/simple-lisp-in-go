package main

import (
    `fmt`
    `math/cmplx`
)

type Intrinsic struct {
    Name string
    Proc func(*Scope, []Value) Value
}

var (
	intrinsicsTab = make(map[string]*Intrinsic)
)

func newIntrinsic(name string, proc func(*Scope, []Value) Value) *Intrinsic {
    return &Intrinsic {
        Name: name,
        Proc: proc,
    }
}

func RegisterIntrinsic(name string, proc func(*Scope, []Value) Value) {
    if _, ok := intrinsicsTab[name]; ok {
        panic("registry: duplicated intrinsic proc: " + name)
    } else {
        intrinsicsTab[name] = newIntrinsic(name, proc)
    }
}

func (self *Intrinsic) Call(sc *Scope, args []Value) Value {
    return self.Proc(sc, args)
}

func (self *Intrinsic) String() string {
    return fmt.Sprintf("#[intrinsic-%s]", self.Name)
}

func (self *Intrinsic) IsIdentity() bool {
    return true
}

/** Arithmetic Operators **/

func intrinsicsAdd(_ *Scope, args []Value) Value {
    return reduceSeq(args, MakeInt(0), func(a Value, b Value) Value {
        switch maxtype(numtypeof(a), numtypeof(b)) {
            case _T_int     : return a.(Int).Add(b.(Int))
            case _T_frac    : return numasfrac(a).Add(numasfrac(b))
            case _T_float   : return numasfloat(a) + numasfloat(b)
            case _T_complex : return numascomplex(a) + numascomplex(b)
            default         : panic("+: unreachable")
        }
    })
}

func intrinsicsSub(_ *Scope, args []Value) Value {
    return reduceSeq(args, MakeInt(0), func(a Value, b Value) Value {
        switch maxtype(numtypeof(a), numtypeof(b)) {
            case _T_int     : return a.(Int).Sub(b.(Int))
            case _T_frac    : return numasfrac(a).Sub(numasfrac(b))
            case _T_float   : return numasfloat(a) - numasfloat(b)
            case _T_complex : return numascomplex(a) - numascomplex(b)
            default         : panic("-: unreachable")
        }
    })
}

func intrinsicsMul(_ *Scope, args []Value) Value {
    return reduceSeq(args, MakeInt(1), func(a Value, b Value) Value {
        switch maxtype(numtypeof(a), numtypeof(b)) {
            case _T_int     : return a.(Int).Mul(b.(Int))
            case _T_frac    : return numasfrac(a).Mul(numasfrac(b))
            case _T_float   : return numasfloat(a) * numasfloat(b)
            case _T_complex : return numascomplex(a) * numascomplex(b)
            default         : panic("*: unreachable")
        }
    })
}

func intrinsicsDiv(_ *Scope, args []Value) Value {
    return reduceSeq(args, MakeInt(1), func(a Value, b Value) Value {
        switch maxtype(numtypeof(a), numtypeof(b)) {
            case _T_int     : return MakeFrac(a.(Int), b.(Int))
            case _T_frac    : return numasfrac(a).Div(numasfrac(b))
            case _T_float   : return numasfloat(a) / numasfloat(b)
            case _T_complex : return numascomplex(a) / numascomplex(b)
            default         : panic("/: unreachable")
        }
    })
}

func init() {
    RegisterIntrinsic("+", intrinsicsAdd)
    RegisterIntrinsic("-", intrinsicsSub)
    RegisterIntrinsic("*", intrinsicsMul)
    RegisterIntrinsic("/", intrinsicsDiv)
}

/** Comparison Operators **/

func intrinsicsEq(_ *Scope, args []Value) Value {
    return Bool(reduceMonotonic(args, func(a Value, b Value) bool {
        switch maxtype(numtypeof(a), numtypeof(b)) {
            case _T_int     : return a.(Int).Cmp(b.(Int)) == 0
            case _T_frac    : return numasfrac(a).Cmp(numasfrac(b)) == 0
            case _T_float   : return numasfloat(a) == numasfloat(b)
            case _T_complex : return numascomplex(a) == numascomplex(b)
            default         : panic("=: unreachable")
        }
    }))
}

func intrinsicsLt(_ *Scope, args []Value) Value {
    return Bool(reduceMonotonic(args, func(a Value, b Value) bool {
        switch maxtype(numtypeof(a), numtypeof(b)) {
            case _T_int     : return a.(Int).Cmp(b.(Int)) < 0
            case _T_frac    : return numasfrac(a).Cmp(numasfrac(b)) < 0
            case _T_float   : return numasfloat(a) < numasfloat(b)
            case _T_complex : panic("<: complex numbers can only be compared for equality")
            default         : panic("<: unreachable")
        }
    }))
}

func intrinsicsGt(_ *Scope, args []Value) Value {
    return Bool(reduceMonotonic(args, func(a Value, b Value) bool {
        switch maxtype(numtypeof(a), numtypeof(b)) {
            case _T_int     : return a.(Int).Cmp(b.(Int)) > 0
            case _T_frac    : return numasfrac(a).Cmp(numasfrac(b)) > 0
            case _T_float   : return numasfloat(a) > numasfloat(b)
            case _T_complex : panic(">: complex numbers can only be compared for equality")
            default         : panic(">: unreachable")
        }
    }))
}

func intrinsicsLte(_ *Scope, args []Value) Value {
    return Bool(reduceMonotonic(args, func(a Value, b Value) bool {
        switch maxtype(numtypeof(a), numtypeof(b)) {
            case _T_int     : return a.(Int).Cmp(b.(Int)) <= 0
            case _T_frac    : return numasfrac(a).Cmp(numasfrac(b)) <= 0
            case _T_float   : return numasfloat(a) <= numasfloat(b)
            case _T_complex : panic("<=: complex numbers can only be compared for equality")
            default         : panic("<=: unreachable")
        }
    }))
}

func intrinsicsGte(_ *Scope, args []Value) Value {
    return Bool(reduceMonotonic(args, func(a Value, b Value) bool {
        switch maxtype(numtypeof(a), numtypeof(b)) {
            case _T_int     : return a.(Int).Cmp(b.(Int)) >= 0
            case _T_frac    : return numasfrac(a).Cmp(numasfrac(b)) >= 0
            case _T_float   : return numasfloat(a) >= numasfloat(b)
            case _T_complex : panic(">=: complex numbers can only be compared for equality")
            default         : panic(">=: unreachable")
        }
    }))
}

func init() {
    RegisterIntrinsic("=", intrinsicsEq)
    RegisterIntrinsic("<", intrinsicsLt)
    RegisterIntrinsic(">", intrinsicsGt)
    RegisterIntrinsic("<=", intrinsicsLte)
    RegisterIntrinsic(">=", intrinsicsGte)
}

/** Unary Arithmetic Functions **/

func intrinsicsRound(_ *Scope, args []Value) Value {
    var nb int
    var vv Value

    /* check for argument count */
    if nb = len(args); nb != 1 {
        panic("round: proc takes exact 1 argument")
    }

    /* round the number */
    switch vv = args[0]; v := vv.(type) {
        case Int     : return vv
        case Frac    : return v.Round()
        case Double  : return v.Round()
        case Complex : return v.Round()
        default      : panic("round: object is not a number: " + AsString(args[0]))
    }
}

func intrinsicsMagnitude(_ *Scope, args []Value) Value {
    if len(args) != 1 {
        panic("magnitude: proc takes exact 1 argument")
    } else if va, ok := args[0].(Complex); !ok {
        panic("magnitude: object is not a complex number: " + AsString(args[0]))
    } else {
        return Double(cmplx.Abs(complex128(va)))
    }
}

func intrinsicsInexactToExact(_ *Scope, args []Value) Value {
    var nb int
    var vv Value

    /* check for argument count */
    if nb = len(args); nb != 1 {
        panic("inexact->exact: proc takes exact 1 argument")
    }

    /* convert the number */
    switch vv = args[0]; v := vv.(type) {
        case Int     : return vv
        case Frac    : return vv
        case Double  : return v.Exact()
        case Complex : return v.Exact()
        default      : panic("inexact->exact: object is not a number: " + AsString(args[0]))
    }
}

func init() {
    RegisterIntrinsic("round", intrinsicsRound)
    RegisterIntrinsic("magnitude", intrinsicsMagnitude)
    RegisterIntrinsic("inexact->exact", intrinsicsInexactToExact)
}

/** Binary Arithmetic Functions **/

func intrinsicsModulo(_ *Scope, args []Value) Value {
    if len(args) != 2 {
        panic("modulo: proc takes exact 2 arguments")
    } else if va, ok := args[0].(Int); !ok {
        panic("modulo: object is not an integer: " + AsString(args[0]))
    } else if vb, ok := args[1].(Int); !ok {
        panic("modulo: object is not an integer: " + AsString(args[1]))
    } else {
        return va.Mod(vb)
    }
}

func intrinsicsQuotient(_ *Scope, args []Value) Value {
    if len(args) != 2 {
        panic("quotient: proc takes exact 2 arguments")
    } else if va, ok := args[0].(Int); !ok {
        panic("quotient: object is not an integer: " + AsString(args[0]))
    } else if vb, ok := args[1].(Int); !ok {
        panic("quotient: object is not an integer: " + AsString(args[1]))
    } else {
        return va.Div(vb)
    }
}

func init() {
    RegisterIntrinsic("modulo", intrinsicsModulo)
    RegisterIntrinsic("quotient", intrinsicsQuotient)
}

/** Value Constructors **/

func intrinsicMakeRectangular(_ *Scope, args []Value) Value {
    if len(args) != 2 {
        panic("make-rectangular: proc takes exact 2 arguments")
    } else {
        return Complex(complex(float64(numasfloat(args[0])), float64(numasfloat(args[1]))))
    }
}

func init() {
    RegisterIntrinsic("make-rectangular", intrinsicMakeRectangular)
}

/** Input / Output Functions **/

func intrinsicsDisplay(_ *Scope, args []Value) Value {
    var ok bool
    var wp *Port

    /* check for arguments */
    if len(args) != 1 && len(args) != 2 {
        panic("display: proc requires 1 or 2 arguments")
    }

    /* check for optional port */
    if wp = PortStdout; len(args) == 2 {
        if wp, ok = args[1].(*Port); !ok {
            panic("display: object is not a port: " + AsString(args[1]))
        }
    }

    /* display the value */
    wp.Write([]byte(AsDisplay(args[0])))
    return nil
}

func intrinsicsNewline(_ *Scope, args []Value) Value {
    var ok bool
    var wp *Port

    /* check for arguments */
    if len(args) != 0 && len(args) != 1 {
        panic("newline: proc requires 1 or 2 arguments")
    }

    /* check for optional port */
    if wp = PortStdout; len(args) == 1 {
        if wp, ok = args[0].(*Port); !ok {
            panic("newline: object is not a port: " + AsString(args[0]))
        }
    }

    /* display the newline */
    wp.Write([]byte{'\n'})
    return nil
}

func intrinsicsCallWithOutputFile(sc *Scope, args []Value) Value {
    var ok bool
    var cb *Proc
    var fn String

    /* extract the file name and callback */
    if len(args) != 2                 { panic("call-with-output-file: proc requires exact 2 arguments") }
    if cb, ok = args[1].(*Proc) ; !ok { panic("call-with-output-file: object is not a proc: " + AsString(args[1])) }
    if fn, ok = args[0].(String); !ok { panic("call-with-output-file: object is not a string: " + AsString(args[0])) }

    /* open a new port */
    file := string(fn)
    port := OpenFileWritePort(file)

    /* call the function with the port */
    defer port.Close()
    return cb.Call(sc, []Value{port})
}

func init() {
    RegisterIntrinsic("display", intrinsicsDisplay)
    RegisterIntrinsic("newline", intrinsicsNewline)
    RegisterIntrinsic("call-with-output-file", intrinsicsCallWithOutputFile)
}
