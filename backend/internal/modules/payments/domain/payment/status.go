package payment

type status int

const (
	StatusUnknown status = iota
	StatusPending
	StatusConfirmed
	StatusCancelled
)

func (s status) Int() int {
	return int(s)
}

type Status struct {
	status status
}

func (s Status) Value() status {
	return s.status
}

func NewStatus(status status) Status {
	if status < StatusUnknown || status > StatusCancelled {
		status = StatusUnknown
	}

	return Status{
		status: status,
	}
}

func FromStatus(raw int) Status {
	status := status(raw)
	if status < StatusUnknown || status > StatusCancelled {
		status = StatusUnknown
	}

	return Status{
		status: status,
	}
}
