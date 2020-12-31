package exit

// Exit code and message.  The root command will look at these and
// log the error and exit with the code when a command completes
var (
	Code int
	Func func()
)

func AddFunc(f func()) {
	g := Func
	Func = func() {
		f()
		if g != nil {
			g()
		}
	}
}
