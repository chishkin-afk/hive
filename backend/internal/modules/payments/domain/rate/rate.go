package rate

import "github.com/google/uuid"

type Rate struct {
	id          uuid.UUID
	name        string
	amountPenny uint64
	paidDays    uint32
}

func From(
	id uuid.UUID,
	name string,
	amountPenny uint64,
	paidDays uint32,
) *Rate {
	return &Rate{
		id:          id,
		name:        name,
		amountPenny: amountPenny,
		paidDays:    paidDays,
	}
}

func (r *Rate) ID() uuid.UUID {
	return r.id
}

func (r *Rate) Name() string {
	return r.name
}

func (r *Rate) AmountPenny() uint64 {
	return r.amountPenny
}

func (r *Rate) PaidDays() uint32 {
	return r.paidDays
}
