package esbapi

type Fn func(m Model) error

type FnGen func(params []interface{}) (Fn, error)
