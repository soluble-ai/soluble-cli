package exit

// Exit code and message.  The root command will look at these and
// log the error and exit with the code when a command completes
var (
	Code    int
	Message string
)
