package models

import "errors"

func (o *Order) Validate() error {
	if o.OrderUID == "" {
		return errors.New("order_uid is required")
	}
	if o.TrackNumber == "" {
		return errors.New("track_number is required")
	}
	if o.CustomerID == "" {
		return errors.New("customer_id is required")
	}
	if o.Delivery.Name == "" {
		return errors.New("delivery.name is required")
	}
	if o.Payment.Transaction == "" {
		return errors.New("payment.transaction is required")
	}
	return nil
}
