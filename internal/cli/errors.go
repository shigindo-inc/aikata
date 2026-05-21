package cli

// ExitError carries a non-zero process exit code alongside an underlying
// error. The cmd/aikata/main entry point inspects this with errors.As to
// pick the right os.Exit value. Allowed codes (per ARCHITECTURE.md §7.2):
//
//	0  — success (no error)
//	1  — generic runtime error (default for any non-ExitError error)
//	2  — user input error (bad flag, missing arg, target dir not empty)
//	3  — aikata doctor found errors (future)
//
// Subcommands return *ExitError when they want anything other than 1.
type ExitError struct {
	Code int
	Err  error
}

// Error reports the wrapped error string so cobra prints it normally.
func (e *ExitError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

// Unwrap exposes the wrapped error for errors.Is / errors.As traversal.
func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
