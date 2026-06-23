package payment

type status int

const (
	StatusUnknown status = iota
	StatusPending
	StatusConfirmed
	StatusCanceled
)

func (s status) Int() int {
	return int(s)
}

func (s status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusConfirmed:
		return "confirmed"
	case StatusCanceled:
		return "canceled"
	}

	return "unknown"
}

type Status struct {
	status status
}

func (s Status) Value() status {
	return s.status
}

func NewStatus(status status) Status {
	if status < StatusUnknown || status > StatusCanceled {
		status = StatusUnknown
	}

	return Status{
		status: status,
	}
}

func FromStatus(raw int) Status {
	status := status(raw)
	if status < StatusUnknown || status > StatusCanceled {
		status = StatusUnknown
	}

	return Status{
		status: status,
	}
}
