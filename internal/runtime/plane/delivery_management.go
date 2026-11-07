package plane

import "context"

type DeliveryManagement struct {
	publisher *DurablePublisher
	store     DeliveryStore
}
type DeliveryView struct {
	Key            string
	Status         DeliveryStatus
	Attempts       int
	LastError      string
	UpdatedAt      string
	OwnerConfirmed bool
}

func NewDeliveryManagement(p *DurablePublisher, s DeliveryStore) *DeliveryManagement {
	return &DeliveryManagement{publisher: p, store: s}
}
func (m *DeliveryManagement) List(ctx context.Context) ([]DeliveryView, error) {
	var rows []*DeliveryAttempt
	var err error
	if all, ok := m.store.(DeliveryAuditStore); ok {
		rows, err = all.ListAll(ctx)
	} else {
		rows, err = m.store.ListRecoverable(ctx)
	}
	if err != nil {
		return nil, err
	}
	out := make([]DeliveryView, 0, len(rows))
	for _, r := range rows {
		out = append(out, deliveryView(r))
	}
	return out, nil
}
func (m *DeliveryManagement) Get(ctx context.Context, key string) (DeliveryView, error) {
	r, err := m.store.Get(ctx, key)
	if err != nil {
		return DeliveryView{}, err
	}
	return deliveryView(r), nil
}
func (m *DeliveryManagement) ConfirmRetry(ctx context.Context, key, owner string) error {
	return m.publisher.ConfirmAndRetry(ctx, key, owner)
}
func deliveryView(r *DeliveryAttempt) DeliveryView {
	return DeliveryView{Key: r.Key, Status: r.Status, Attempts: r.Attempts, LastError: r.LastError, UpdatedAt: r.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"), OwnerConfirmed: r.OwnerConfirmed}
}
