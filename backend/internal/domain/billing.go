package domain

import (
	"time"
)

type Billing struct {
	ID          string    `json:"id"`
	PatientID   string    `json:"patient_id"`
	EncounterID string    `json:"encounter_id"`
	Total       int       `json:"total"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Invoice struct {
	ID        string     `json:"id"`
	BillingID string     `json:"billing_id"`
	Amount    int        `json:"amount"`
	PaidAt    *time.Time `json:"paid_at"`
}

type Payment struct {
	ID        string    `json:"id"`
	BillingID string    `json:"billing_id"`
	Amount    int       `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type BPJSSEP struct {
	NoSEP       string    `json:"no_sep"`
	PatientID   string    `json:"patient_id"`
	EncounterID string    `json:"encounter_id"`
	Polyclinic  string    `json:"polyclinic"`
	Total       int       `json:"total"`
	CreatedAt   time.Time `json:"created_at"`
}

type BPJSClaim struct {
	ID        string    `json:"id"`
	NoSEP     string    `json:"no_sep"`
	PatientID string    `json:"patient_id"`
	Status    string    `json:"status"`
	Amount    int       `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type FinancialReport struct {
	TotalRevenue int `json:"total_revenue"`
	TotalPaid    int `json:"total_paid"`
	TotalPending int `json:"total_pending"`
	TotalBPJS    int `json:"total_bpjs"`
}
