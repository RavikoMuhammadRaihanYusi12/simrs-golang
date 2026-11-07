package usecase

import (
	"context"
	"testing"

	"github.com/thinkingx/simrs-backend/internal/domain"
)

type MockBPJSStore struct {
	claims []domain.BPJSClaim
}

func (m *MockBPJSStore) CreateClaim(ctx context.Context, c domain.BPJSClaim) error {
	m.claims = append(m.claims, c)
	return nil
}

func (m *MockBPJSStore) GetClaimStatus(ctx context.Context, noSEP string) (*domain.BPJSClaim, error) {
	for i := range m.claims {
		if m.claims[i].NoSEP == noSEP {
			return &m.claims[i], nil
		}
	}
	return nil, nil
}

func TestBPJSUsecase_CreateSEP(t *testing.T) {
	store := &MockBPJSStore{}
	uc := NewBPJSUsecase(store)

	sep, err := uc.CreateSEP(context.Background(), "patient1", "enc1", "Poli Umum", 150000)
	if err != nil || sep == nil {
		t.Fatal("expected sep, got error:", err)
	}
	if sep.NoSEP == "" || sep.PatientID != "patient1" || sep.Total != 150000 {
		t.Fatal("bad sep fields:", sep)
	}
	if len(store.claims) != 1 || store.claims[0].Status != "submitted" {
		t.Fatal("claim not recorded correctly")
	}
}

func TestBPJSUsecase_GetClaimStatus(t *testing.T) {
	store := &MockBPJSStore{}
	uc := NewBPJSUsecase(store)

	sep, _ := uc.CreateSEP(context.Background(), "p1", "e1", "Poli", 1000)
	claim, err := uc.GetClaimStatus(context.Background(), sep.NoSEP)
	if err != nil || claim == nil || claim.NoSEP != sep.NoSEP {
		t.Fatal("claim lookup failed")
	}
}
