package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/thinkingx/simrs-backend/internal/delivery"
	"github.com/thinkingx/simrs-backend/internal/repository"
	"github.com/thinkingx/simrs-backend/internal/usecase"
	"github.com/thinkingx/simrs-backend/pkg/jwt"
)

func main() {
	dsn := "file:./simrs.db?mode=memory&cache=shared"

	patientRepo, err := repository.NewSQLitePatientStore(dsn)
	if err != nil {
		log.Fatalf("Failed to open patient DB: %v", err)
	}
	defer patientRepo.Close()

	medicalRepo, err := repository.NewSQLiteMedicalRepository(dsn)
	if err != nil {
		log.Fatalf("Failed to open medical DB: %v", err)
	}
	defer medicalRepo.Close()

	billingRepo, err := repository.NewSQLiteBillingRepository(dsn)
	if err != nil {
		log.Fatalf("Failed to open billing DB: %v", err)
	}
	defer billingRepo.Close()

	patientUC := usecase.NewPatientUsecase(patientRepo)
	medicalUC := usecase.NewMedicalUsecase(medicalRepo)
	billingUC := usecase.NewBillingUsecase(billingRepo)
	bpjsUC := usecase.NewBPJSUsecase(billingRepo)

	jwtSvc := jwt.NewJWTService("secret-key-change-prod")
	hdl := delivery.NewDelivery(patientUC, medicalUC, billingUC, bpjsUC, jwtSvc)

	mux := hdl.RegisterRoutes()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("SIMRS backend starting on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server error: %v", err)
	}
}
