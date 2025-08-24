package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tmozzze/order_checker/internal/models"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) GetOrderById(ctx context.Context, orderID string) (*models.Order, error) {
	// Get order
	query := `
		SELECT order_uid, track_number, enrty, locale, internal_signature, customer_id,
			delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1
	`

	row := r.pool.QueryRow(ctx, query, orderID)

	var o models.Order
	err := row.Scan(
		&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature,
		&o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated,
		&o.OofShard,
	)
	if err != nil {
		return nil, err
	}

	// Get Delivery

	deliveryQuery := `
		SELECT name, phone, zip, city, address, region, email
		FROM delivery WHERE order_uid = $1
	`
	row = r.pool.QueryRow(ctx, deliveryQuery, orderID)

	var d models.Delivery
	err = row.Scan(&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email)
	if err != nil {
		return nil, err
	}

	// Delivery in Order struct
	o.Delivery = d

	// Get Payment
	paymentDelivery := `
		SELECT transaction, request_id, currency, provider, amount,
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payments WHERE order_uid = $1
	`
	row = r.pool.QueryRow(ctx, paymentDelivery, orderID)

	var p models.Payment
	err = row.Scan(&p.Transaction, &p.RequestID, &p.Currency, &p.Provider,
		&p.Amount, &p.PaymentDt, &p.Bank, &p.DeliveryCost, &p.GoodsTotal,
		&p.CustomFee,
	)
	if err != nil {
		return nil, err
	}

	// Payment in Order struct
	o.Payment = p

	// Get Items
	queryItems := `
		SELECT chrt_id, track_number, price, rid, name, sale, size,
			total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1
	`

	rows, err := r.pool.Query(ctx, queryItems, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item

		err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand,
			&item.Status,
		)
		if err != nil {
			return nil, err
		}

		// Item to slice Items in Order struct
		o.Items = append(o.Items, item)
	}

	return &o, nil

}
