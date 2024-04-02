package goerror

type GoError struct {
	err     error
	Errfunc ErrorFuc
}

func Default() *GoError {
	return &GoError{}
}
func (e *GoError) Error() string {
	return e.err.Error()
}

func (e *GoError) Put(err error) {
	e.check(err)
}

func (e *GoError) check(err error) {
	if err != nil {
		e.err = err
		panic(e)
	}
}

type ErrorFuc func(msError *GoError)

func (e *GoError) Result(errFuc ErrorFuc) {
	e.Errfunc = errFuc
}
func (e *GoError) ExecResult() {
	e.Errfunc(e)
}
