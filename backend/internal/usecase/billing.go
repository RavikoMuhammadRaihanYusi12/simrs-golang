package usecase

import (
	"context"
	"errors"

	"github.com/thinkingx/simrs-backend/internal/domain"
)

var (
	ErrBillingNotFound = errors.New("billing not found")
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrAlreadyPaid     = errors.New("already paid")
)

type BillingStore interface {
	CreateBilling(ctx context.Context, b domain.Billing) error
	GetBilling(ctx context.Context, id string) (*domain.Billing, error)
	GetBillingsByPatient(ctx context.Context, patientID string) ([]domain.Billing, error)
	GetBillingsByEncounter(ctx context.Context, encounterID string) ([]domain.Billing, error)
	UpdateBillingStatus(ctx context.Context, id string, status string) error
	CreatePayment(ctx context.Context, p domain.Payment) error
	GetPaymentsByBilling(ctx context.Context, billingID string) ([]domain.Payment, error)
}

type BillingUsecase struct {
	store BillingStore
}

func NewBillingUsecase(store BillingStore) *BillingUsecase {
	return &BillingUsecase{store: store}
}

func (u *BillingUsecase) CreateBilling(ctx context.Context, patientID, encounterID string, amount int) (*domain.Billing, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	b := domain.Billing{
		PatientID:   patientID,
		EncounterID: encounterID,
		Total:       amount,
		Status:      "pending",
	}
	if err := u.store.CreateBilling(ctx, b); err != nil {
		return nil, err
	}
	return &b, nil
}

func (u *BillingUsecase) GetBilling(ctx context.Context, id string) (*domain.Billing, error) {
	b, err := u.store.GetBilling(ctx, id)
	if err != nil {
		return nil, ErrBillingNotFound
	}
	return b, nil
}

func (u *BillingUsecase) PayBilling(ctx context.Context, billingID string, amount int) error {
	b, err := u.store.GetBilling(ctx, billingID)
	if err != nil {
		return ErrBillingNotFound
	}
	if b.Status == "paid" {
		return ErrAlreadyPaid
	}
	if amount != b.Total {
		return ErrInvalidAmount
	}
	p := domain.Payment{
		BillingID: billingID,
		Amount:    amount,
	}
	if err := u.store.CreatePayment(ctx, p); err != nil {
		return err
	}
	return u.store.UpdateBillingStatus(ctx, billingID, "paid")
}

func (u *BillingUsecase) GetBillingsByPatient(ctx context.Context, patientID string) ([]domain.Billing, error) {
	return u.store.GetBillingsByPatient(ctx, patientID)
}

func (u *BillingUsecase) GetBillingsByEncounter(ctx context.Context, encounterID string) ([]domain.Billing, error) {
	return u.store.GetBillingsByEncounter(ctx, encounterID)
}

func (u *BillingUsecase) UpdateBillingStatus(ctx context.Context, id string, status string) error {
	return u.store.UpdateBillingStatus(ctx, id, status)
}

func (u *BillingUsecase) GetPaymentsByBilling(ctx context.Context, billingID string) ([]domain.Payment, error) {
	return u.store.GetPaymentsByBilling(ctx, billingID)
}
