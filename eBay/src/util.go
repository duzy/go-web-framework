package eBay

import (
        "os"
        "fmt"
        "strconv"
        "reflect"
        //"runtime"
)

// AssignFields assign left-hand-side fields to right-hand-side fields of
// same names.
func AssignFields_(lhs, rhs interface{}) (err os.Error) {
        var rv *reflect.StructValue
        if v, ok := reflect.NewValue(rhs).(*reflect.PtrValue).Elem().(*reflect.StructValue); ok {
                rv = v
        } else {
                err = os.NewError("args is not a *StructValue")
        }

        if err != nil { return }

        lv := reflect.NewValue(lhs).(*reflect.PtrValue).Elem()
        lt := lv.Type().(*reflect.StructType)
        for i := 0; i < lt.NumField(); i += 1 {
                var lft reflect.StructField

                if lft = lt.Field(i); lft.Anonymous {
                        // FIXME: bind base-struct fields?
                        continue
                }

                if rfv := rv.FieldByName(lft.Name); rfv != nil {
                        lfv := lv.(*reflect.StructValue).FieldByIndex(lft.Index)
                        lfv.SetValue(rfv)
                }
        }
        return
}

func AssignFields(lhs, rhs interface{}) (err os.Error) {
        err = MapFields(lhs, rhs, func(l, r reflect.Value)(nxt bool) {
                l.SetValue(r)
                return true
        })
        return
}

func ForEachField(s interface{}, f func(t *reflect.StructField, v reflect.Value)(nxt bool)) (err os.Error) {
        var ok bool
        var v *reflect.StructValue
        switch p := reflect.NewValue(s).(type) {
        case *reflect.StructValue: v, ok = p, true
        case *reflect.PtrValue: v, ok = p.Elem().(*reflect.StructValue)
        }

        if !ok { err = os.NewError("interface is not *reflect.StructValue"); return }

        t, ok := v.Type().(*reflect.StructType)
        for i := 0; i < t.NumField(); i += 1 {
                ft := t.Field(i)
                fv := v.FieldByIndex(ft.Index)

                if ft.Anonymous {
                        //ForEachField(fv.Interface(), f)
                        continue
                }

                if !f(&ft, fv) { break }
        }
        return
}

// MapFields invoke 'f' for each field of 'lhs' occurs in 'rhs'
func MapFields(lhs, rhs interface{}, f func(lf, rf reflect.Value)(nxt bool)) (err os.Error) {
        lv, ok := reflect.NewValue(lhs).(*reflect.PtrValue).Elem().(*reflect.StructValue)
        if !ok { err = os.NewError("lhs is not *reflect.StructValue"); return }
        lt, ok := lv.Type().(*reflect.StructType)

        rv, ok := reflect.NewValue(rhs).(*reflect.PtrValue).Elem().(*reflect.StructValue)
        if !ok { err = os.NewError("rhs is not *reflect.StructValue"); return }
        rt, ok := rv.Type().(*reflect.StructType)

        for i := 0; i < lt.NumField(); i += 1 {
                var t1, t2 reflect.StructField
                if t1 = lt.Field(i); t1.Anonymous {
                        // FIXME: ...
                        continue
                }

                if t2, ok = rt.FieldByName(t1.Name); !ok {
                        // TODO: handle with Anonymous?
                        continue
                }

                //if rf := rv.FieldByName(t1.Name); rf != nil {
                if rf := rv.FieldByIndex(t2.Index); rf != nil {
                        v := lv.FieldByIndex(t1.Index)
                        if !f(v, rf) { break }
                }
        }
        return
}

func ConvertValue(k reflect.Kind, v reflect.Value) (ov reflect.Value) {
        if k == v.Type().Kind() { ov = v; return }

        if a, ok := v.(reflect.ArrayOrSliceValue); ok {
                if 0 < a.Len() { v = a.Elem(0) } else { return }
        }

        if k == v.Type().Kind() { ov = v; return }

        s := v.Interface().(string) // TODO: arbitray type
        switch k {
        case reflect.Bool:      if o, e := strconv.Atob(s); e == nil { ov = reflect.NewValue(o) }
        case reflect.Int:       if o, e := strconv.Atoi(s); e == nil { ov = reflect.NewValue(o) }
        case reflect.Float:     if o, e := strconv.Atof(s); e == nil { ov = reflect.NewValue(o) }
        case reflect.String:    ov = reflect.NewValue(s)
        default:
                fmt.Printf("todo: convert: (%s) %v -> (%s)\n", v.Type().Kind(), v.Interface(), k)
        }
        return
}

func RoughAssignValue(lhs, rhs reflect.Value) (err os.Error) {
        //if p, ok := lhs.(*reflect.InterfaceValue); ok { lhs = p.Elem(); }
        //if p, ok := rhs.(*reflect.InterfaceValue); ok { rhs = p.Elem(); }
        if p, ok := lhs.(*reflect.PtrValue); ok { lhs = p.Elem(); }
        if p, ok := rhs.(*reflect.PtrValue); ok { rhs = p.Elem(); }

        switch lv := lhs.(type) {
        case *reflect.StructValue:
                // Make sure rhs is also StructValue
                //      eg: []struct{} -> struct{}
                rhs = ConvertValue(reflect.Struct, rhs)
                if rhs == nil {
                        err = os.NewError("rhs is not *reflect.StructValue")
                }

                if rv, ok := rhs.(*reflect.StructValue); ok {
                        lt := lv.Type().(*reflect.StructType)
                        //rt := rv.Type().(*reflect.StructType)
                        for i := 0; i < lt.NumField(); i += 1 {
                                ft := lt.Field(i)
                                fv := lv.FieldByIndex(ft.Index)
                                if v := rv.FieldByName(ft.Name); v != nil {
                                        err = RoughAssignValue(fv, v)
                                }
                        }
                } else {
                        err = os.NewError("rhs is not *reflect.StructValue")
                }
        case reflect.ArrayOrSliceValue:
                if rv, ok := rhs.(reflect.ArrayOrSliceValue); ok {
                        //lv = reflect.MakeSlice(lv.Type().(*reflect.SliceType), rv.Len(), rv.Len())
                        for i := 0 ; i < lv.Len() && i < rv.Len(); i += 1 {
                                err = RoughAssignValue(lv.Elem(i), rv.Elem(i))
                        }
                } else {
                        //lv = reflect.MakeSlice(lv.Type().(*reflect.SliceType), 1, 1)
                        if 0 < lv.Len() {
                                err = RoughAssignValue(lv.Elem(0), rv)
                        }
                }
        default:
                if v := ConvertValue(lhs.Type().Kind(), rhs); v != nil {
                        lhs.SetValue(v)
                } else {
                        fmt.Printf("todo: assign: (%s) = (%s) %v\n", lhs.Type().Kind(), rhs.Type().Kind(), rhs.Interface())
                }
        }
        return
}

func RoughAssign(lhs, rhs interface{}) (err os.Error) {
        defer func() {
                if r := recover(); r != nil {
                        switch e := r.(type) {
                        case os.Error:
                                err = e
                        case string:
                                err = os.NewError(e)
                        default:
                                panic(r)
                        }
                }
        }()

        lv, rv := reflect.NewValue(lhs), reflect.NewValue(rhs)
        return RoughAssignValue(lv, rv)
}
