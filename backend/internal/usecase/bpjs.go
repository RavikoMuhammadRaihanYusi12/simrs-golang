package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/thinkingx/simrs-backend/internal/domain"
)

type BPJSStore interface {
	CreateClaim(ctx context.Context, c domain.BPJSClaim) error
	GetClaimStatus(ctx context.Context, noSEP string) (*domain.BPJSClaim, error)
}

type BPJSUsecase struct {
	store BPJSStore
}

func NewBPJSUsecase(store BPJSStore) *BPJSUsecase {
	return &BPJSUsecase{store: store}
}

func (u *BPJSUsecase) CreateSEP(ctx context.Context, patientID, encounterID, polyclinic string, amount int) (*domain.BPJSSEP, error) {
	noSEP := fmt.Sprintf("SEP-%d", time.Now().UnixNano())
	sep := &domain.BPJSSEP{
		NoSEP:       noSEP,
		PatientID:   patientID,
		EncounterID: encounterID,
		Polyclinic:  polyclinic,
		Total:       amount,
		CreatedAt:   time.Now(),
	}
	claim := &domain.BPJSClaim{
		ID:        noSEP,
		NoSEP:     noSEP,
		PatientID: patientID,
		Status:    "submitted",
		Amount:    amount,
		CreatedAt: time.Now(),
	}
	if err := u.store.CreateClaim(ctx, *claim); err != nil {
		return nil, err
	}
	return sep, nil
}

func (u *BPJSUsecase) GetClaimStatus(ctx context.Context, noSEP string) (*domain.BPJSClaim, error) {
	return u.store.GetClaimStatus(ctx, noSEP)
}
