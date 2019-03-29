// errorshandler project errorshandler.go
package errorshandler

func ErrorHandle(e error) {
	if e != nil {
		panic(e)
	}
}
