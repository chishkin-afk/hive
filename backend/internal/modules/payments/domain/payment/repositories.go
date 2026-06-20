package payment

import "context"

type UpdateFunc func(ctx context.Context, payment *Payment) error
