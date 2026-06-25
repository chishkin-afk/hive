package rateredis

import (
	"encoding/json"

	"github.com/devathh/hive/backend/internal/modules/payments/domain/rate"
)

func rateToBytes(rate *rate.Rate) ([]byte, error) {
	return json.Marshal(rateRecord{
		ID:          rate.ID(),
		Name:        rate.Name(),
		AmountPenny: rate.AmountPenny(),
		PaidDays:    rate.PaidDays(),
	})
}

func bytesToRate(bytes []byte) (*rate.Rate, error) {
	var record rateRecord
	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

	return rate.From(
		record.ID,
		record.Name,
		record.AmountPenny,
		record.PaidDays,
	), nil
}

func ratesToBytes(rates []*rate.Rate) ([]byte, error) {
	records := make([]rateRecord, len(rates))
	for i, rate := range rates {
		records[i] = rateRecord{
			ID:          rate.ID(),
			Name:        rate.Name(),
			AmountPenny: rate.AmountPenny(),
			PaidDays:    rate.PaidDays(),
		}
	}

	return json.Marshal(records)
}

func bytesToRates(bytes []byte) ([]*rate.Rate, error) {
	var records []rateRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
		return nil, err
	}

	rates := make([]*rate.Rate, len(records))
	for i, record := range records {
		rates[i] = rate.From(
			record.ID,
			record.Name,
			record.AmountPenny,
			record.PaidDays,
		)
	}

	return rates, nil
}
