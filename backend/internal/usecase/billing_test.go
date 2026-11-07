package usecase

import (
	"context"
	"testing"

	"github.com/thinkingx/simrs-backend/internal/domain"
)

type MockBillingStore struct {
	CreateBillingFunc          func(ctx context.Context, b domain.Billing) error
	GetBillingFunc             func(ctx context.Context, id string) (*domain.Billing, error)
	GetBillingsByPatientFunc   func(ctx context.Context, patientID string) ([]domain.Billing, error)
	GetBillingsByEncounterFunc func(ctx context.Context, encounterID string) ([]domain.Billing, error)
	UpdateBillingStatusFunc    func(ctx context.Context, id string, status string) error
	CreatePaymentFunc          func(ctx context.Context, p domain.Payment) error
	GetPaymentsByBillingFunc   func(ctx context.Context, billingID string) ([]domain.Payment, error)
	CloseFunc                  func() error
}

func (m *MockBillingStore) CreateBilling(ctx context.Context, b domain.Billing) error {
	return m.CreateBillingFunc(ctx, b)
}
func (m *MockBillingStore) GetBilling(ctx context.Context, id string) (*domain.Billing, error) {
	return m.GetBillingFunc(ctx, id)
}
func (m *MockBillingStore) GetBillingsByPatient(ctx context.Context, patientID string) ([]domain.Billing, error) {
	return m.GetBillingsByPatientFunc(ctx, patientID)
}
func (m *MockBillingStore) GetBillingsByEncounter(ctx context.Context, encounterID string) ([]domain.Billing, error) {
	return m.GetBillingsByEncounterFunc(ctx, encounterID)
}
func (m *MockBillingStore) UpdateBillingStatus(ctx context.Context, id string, status string) error {
	return m.UpdateBillingStatusFunc(ctx, id, status)
}
func (m *MockBillingStore) CreatePayment(ctx context.Context, p domain.Payment) error {
	return m.CreatePaymentFunc(ctx, p)
}
func (m *MockBillingStore) GetPaymentsByBilling(ctx context.Context, billingID string) ([]domain.Payment, error) {
	return m.GetPaymentsByBillingFunc(ctx, billingID)
}
func (m *MockBillingStore) Close() error {
	return m.CloseFunc()
}

func TestBillingUsecase_CreateBilling(t *testing.T) {
	store := &MockBillingStore{}
	uc := NewBillingUsecase(store)

	t.Run("success", func(t *testing.T) {
		store.CreateBillingFunc = func(ctx context.Context, b domain.Billing) error {
			return nil
		}

		b, err := uc.CreateBilling(context.Background(), "patient1", "enc1", 1000)
		if err != nil || b == nil || b.Total != 1000 || b.Status != "pending" {
			t.Fatal("fail")
		}
	})

	t.Run("invalid amount", func(t *testing.T) {
		b, err := uc.CreateBilling(context.Background(), "p", "e", -5)
		if err != ErrInvalidAmount || b != nil {
			t.Fatal("fail")
		}
	})
}

func TestBillingUsecase_PayBilling(t *testing.T) {
	store := &MockBillingStore{}
	uc := NewBillingUsecase(store)

	t.Run("success", func(t *testing.T) {
		store.GetBillingFunc = func(ctx context.Context, id string) (*domain.Billing, error) {
			return &domain.Billing{ID: id, Total: 500, Status: "pending"}, nil
		}
		store.CreatePaymentFunc = func(ctx context.Context, p domain.Payment) error {
			return nil
		}
		store.UpdateBillingStatusFunc = func(ctx context.Context, id string, status string) error {
			return nil
		}

		if err := uc.PayBilling(context.Background(), "bill1", 500); err != nil {
			t.Fatal("fail")
		}
	})

	t.Run("already paid", func(t *testing.T) {
		store.GetBillingFunc = func(ctx context.Context, id string) (*domain.Billing, error) {
			return &domain.Billing{ID: id, Total: 300, Status: "paid"}, nil
		}

		if err := uc.PayBilling(context.Background(), "bill2", 300); err != ErrAlreadyPaid {
			t.Fatal("fail")
		}
	})
}

func TestBillingUsecase_GetBillingsByPatient(t *testing.T) {
	store := &MockBillingStore{}
	uc := NewBillingUsecase(store)

	store.GetBillingsByPatientFunc = func(ctx context.Context, patientID string) ([]domain.Billing, error) {
		return []domain.Billing{{ID: "b1", Total: 100}}, nil
	}

	billings, err := uc.GetBillingsByPatient(context.Background(), "patient1")
	if err != nil || len(billings) != 1 || billings[0].ID != "b1" {
		t.Fatal("fail")
	}
}
