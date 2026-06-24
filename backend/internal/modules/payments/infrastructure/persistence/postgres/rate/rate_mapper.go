package ratepg

import "github.com/devathh/hive/backend/internal/modules/payments/domain/rate"

func recordToRate(record *rateRecord) *rate.Rate {
	return rate.From(
		record.ID,
		record.Name,
		record.AmountPenny,
		record.PaidDays,
	)
}

func recordsToRates(records []*rateRecord) []*rate.Rate {
	rates := make([]*rate.Rate, len(records))
	for i, record := range records {
		rates[i] = recordToRate(record)
	}

	return rates
}
