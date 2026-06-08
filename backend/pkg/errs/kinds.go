package errs

type kind int

const (
	KindRequest kind = iota
	KindInternal
	KindUnauthorized
	KindConflict
	KindTimeout
	KindNotFound
	KindPermissionDenied
)

type ErrorKind struct {
	kind kind
	err  error
}

func NewErrorWithKind(err error, kind kind) *ErrorKind {
	return &ErrorKind{
		err:  err,
		kind: kind,
	}
}

func (e *ErrorKind) Error() string {
	return e.err.Error()
}

func (e *ErrorKind) Unwrap() error {
	return e.err
}

func (e *ErrorKind) Kind() kind {
	return e.kind
}
