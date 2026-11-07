package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/thinkingx/simrs-backend/internal/domain"
	"github.com/thinkingx/simrs-backend/internal/usecase"
	"github.com/thinkingx/simrs-backend/pkg/jwt"
)

// cors adds CORS headers to allow the Next.js dev server (port 3000) on
// another origin to call the backend. Allowed origins are configurable so
// production can restrict them.
func cors(allowedOrigins []string) func(http.Handler) http.Handler {
	originSet := make(map[string]bool)
	for _, o := range allowedOrigins {
		originSet[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			// Dev-default: allow localhost origins if none configured.
			allowed := len(originSet) == 0 && origin != "" &&
				(origin == "http://localhost:3000" || origin == "http://127.0.0.1:3000" || origin == "http://localhost:3001")
			if !allowed {
				allowed = originSet[origin]
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type Delivery struct {
	patientUsecase *usecase.PatientUsecase
	medicalUsecase *usecase.MedicalUsecase
	billingUsecase *usecase.BillingUsecase
	bpjsUsecase    *usecase.BPJSUsecase
	jwtService     *jwt.JWTService
}

func NewDelivery(
	patientUsecase *usecase.PatientUsecase,
	medicalUsecase *usecase.MedicalUsecase,
	billingUsecase *usecase.BillingUsecase,
	bpjsUsecase *usecase.BPJSUsecase,
	jwtService *jwt.JWTService,
) *Delivery {
	return &Delivery{
		patientUsecase: patientUsecase,
		medicalUsecase: medicalUsecase,
		billingUsecase: billingUsecase,
		bpjsUsecase:    bpjsUsecase,
		jwtService:     jwtService,
	}
}

// guard wraps a handler requiring a valid JWT whose role is in allow.
// Empty allow = any authenticated user. Login stays unguarded (see RegisterRoutes).
func (d *Delivery) guard(allow []string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS preflight is handled by the outer middleware; let it through.
		if r.Method == http.MethodOptions {
			next(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		claims, err := d.jwtService.Verify(auth[len(prefix):])
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		if len(allow) > 0 && !jwt.RequireAllowlist(claims, allow...) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func (d *Delivery) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Auth (public)
	mux.HandleFunc("/api/v1/auth/login", d.handleLogin)

	// Patients — front office (nurse) + admin; pharmacy reads for dispensing
	mux.HandleFunc("/api/v1/patients", d.guard([]string{"nurse", "admin", "pharmacy"}, d.handlePatients))

	// Master data — any authenticated user can read
	mux.HandleFunc("/api/v1/polyclinics", d.guard(nil, d.handlePolyclinics))
	mux.HandleFunc("/api/v1/doctors", d.guard(nil, d.handleDoctors))
	// Doctor CRUD — admin only (writes inside handler)
	mux.HandleFunc("/api/v1/doctors/", d.guard([]string{"admin"}, d.handleDoctorByID))

	// Encounters / kunjungan — front office creates; doctor, nurse, pharmacy read
	mux.HandleFunc("/api/v1/encounters", d.guard([]string{"nurse", "admin", "doctor", "pharmacy"}, d.handleEncounters))
	mux.HandleFunc("/api/v1/encounters/", d.guard([]string{"nurse", "admin", "doctor", "pharmacy"}, d.handleEncounterByID))

	// Medical records — doctor only
	mux.HandleFunc("/api/v1/medical-records", d.guard([]string{"doctor"}, d.handleMedicalRecords))

	// Pharmacy — pharmacy + admin manage medicines; doctor + pharmacy write prescriptions
	mux.HandleFunc("/api/v1/pharmacy/medicines", d.guard([]string{"pharmacy", "admin"}, d.handleMedicines))
	mux.HandleFunc("/api/v1/pharmacy/medicines/", d.guard([]string{"pharmacy", "admin"}, d.handleMedicineByID))
	mux.HandleFunc("/api/v1/pharmacy/transactions", d.guard([]string{"pharmacy", "admin"}, d.handleMedicineTransactions))
	mux.HandleFunc("/api/v1/pharmacy/prescriptions", d.guard([]string{"doctor", "pharmacy"}, d.handlePrescriptions))
	mux.HandleFunc("/api/v1/pharmacy/prescriptions/", d.guard([]string{"pharmacy", "admin"}, d.handlePrescriptionByID))

	// Billing (Phase 3) — cashier + admin
	mux.HandleFunc("/api/v1/billings", d.guard([]string{"cashier", "admin"}, d.handleBillings))
	mux.HandleFunc("/api/v1/billings/", d.guard([]string{"cashier", "admin"}, d.handleBillingByID))
	mux.HandleFunc("/api/v1/payments", d.guard([]string{"cashier", "admin"}, d.handlePayments))
	mux.HandleFunc("/api/v1/payments/", d.guard([]string{"cashier", "admin"}, d.handlePaymentByID))

	// BPJS (Phase 3) — cashier + admin
	mux.HandleFunc("/api/v1/bpjs/claims", d.guard([]string{"cashier", "admin"}, d.handleBPJSClaims))
	mux.HandleFunc("/api/v1/bpjs/claims/", d.guard([]string{"cashier", "admin"}, d.handleBPJSByNoSEP))

	// CORS: wrap mux so Next.js on another origin can call the API.
	return cors(nil)(mux)
}

func (d *Delivery) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	token, err := d.jwtService.Login(body["username"], body["password"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"token":"` + token + `"}`))
}

func (d *Delivery) handlePatients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		patients, err := d.patientUsecase.ListPatients(r.Context())
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(patients); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
		}
	case http.MethodPost:
		// Pharmacy may read the patient list (for dispensing) but not register.
		if role := d.roleOf(r); role == "pharmacy" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		var p domain.Patient
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := d.patientUsecase.RegisterPatient(r.Context(), p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func (d *Delivery) handlePolyclinics(w http.ResponseWriter, r *http.Request) {
	polys, err := d.patientUsecase.ListPolyclinics(r.Context())
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(polys)
}

func (d *Delivery) handleDoctors(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		poli := r.URL.Query().Get("poli")
		docs, err := d.patientUsecase.ListDoctors(r.Context(), poli)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(docs)

	case http.MethodPost:
		if !d.requireAdmin(w, r) {
			return
		}
		var doc domain.Doctor
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if doc.ID == "" {
			doc.ID = "doc-" + strconv.FormatInt(time.Now().UnixNano(), 36)
		}
		if err := d.patientUsecase.CreateDoctor(r.Context(), doc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(doc)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (d *Delivery) handleDoctorByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/doctors/"):]
	switch r.Method {
	case http.MethodPut:
		var doc domain.Doctor
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		doc.ID = id
		if err := d.patientUsecase.UpdateDoctor(r.Context(), doc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		if err := d.patientUsecase.DeleteDoctor(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// roleOf returns the caller's role from the JWT, or "" if absent/invalid.
func (d *Delivery) roleOf(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(auth) <= len(prefix) {
		return ""
	}
	claims, err := d.jwtService.Verify(auth[len(prefix):])
	if err != nil {
		return ""
	}
	role, _ := claims["role"].(string)
	return role
}

// requireAdmin checks the caller's role; writes 403 and returns false otherwise.
func (d *Delivery) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(auth) <= len(prefix) {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return false
	}
	claims, err := d.jwtService.Verify(auth[len(prefix):])
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return false
	}
	if !jwt.RequireAllowlist(claims, "admin") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (d *Delivery) handleEncounters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var body struct {
			PatientID  string `json:"patient_id"`
			Polyclinic string `json:"polyclinic"`
			DoctorID   string `json:"doctor_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		e, err := d.patientUsecase.CreateVisit(r.Context(), body.PatientID, body.Polyclinic, body.DoctorID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(e)

	case http.MethodGet:
		polyclinic := r.URL.Query().Get("polyclinic")
		encounters, err := d.patientUsecase.ListEncounters(r.Context(), polyclinic)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(encounters)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (d *Delivery) handleEncounterByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/encounters/"):]
	switch r.Method {
	case http.MethodGet:
		e, err := d.patientUsecase.GetEncounter(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(e)

	case http.MethodPatch:
		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := d.patientUsecase.UpdateEncounterStatus(r.Context(), id, body.Status); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (d *Delivery) handleMedicalRecords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var mr domain.MedicalRecord
		if err := json.NewDecoder(r.Body).Decode(&mr); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := d.medicalUsecase.CreateMedicalRecord(r.Context(), mr); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"created"}`))

	case http.MethodGet:
		encounterID := r.URL.Query().Get("encounter_id")
		records, err := d.medicalUsecase.GetMedicalRecords(r.Context(), encounterID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(records); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (d *Delivery) handleMedicines(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var med domain.Medicine
		if err := json.NewDecoder(r.Body).Decode(&med); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := d.medicalUsecase.CreateMedicine(r.Context(), med); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"created"}`))

	case http.MethodGet:
		medicines, err := d.medicalUsecase.ListMedicines(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(medicines); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (d *Delivery) handlePrescriptions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var p domain.Prescription
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := d.medicalUsecase.CreatePrescription(r.Context(), p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"prescribed"}`))

	case http.MethodGet:
		encounterID := r.URL.Query().Get("encounter_id")
		// No encounter_id → pharmacist wants every prescription (dispensing queue).
		if encounterID == "" {
			prescriptions, err := d.medicalUsecase.ListPrescriptions(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(prescriptions)
			return
		}
		prescriptions, err := d.medicalUsecase.GetPrescriptions(r.Context(), encounterID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(prescriptions); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePrescriptionByID moves a prescription status (pending/processing/dispensed).
func (d *Delivery) handlePrescriptionByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/pharmacy/prescriptions/"):]
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := d.medicalUsecase.SetPrescriptionStatus(r.Context(), id, body.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleMedicineByID updates or deletes a single medicine (pharmacy/admin).
func (d *Delivery) handleMedicineByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/pharmacy/medicines/"):]
	switch r.Method {
	case http.MethodPut:
		var med domain.Medicine
		if err := json.NewDecoder(r.Body).Decode(&med); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		med.ID = id
		if err := d.medicalUsecase.UpdateMedicine(r.Context(), med); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		if err := d.medicalUsecase.DeleteMedicine(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMedicineTransactions lists stock transactions (GET) or records a
// take-out (POST) when the pharmacist dispenses stock manually.
func (d *Delivery) handleMedicineTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		trxs, err := d.medicalUsecase.ListMedicineTransactions(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trxs)

	case http.MethodPost:
		var body struct {
			MedicineID string `json:"medicine_id"`
			Quantity   int    `json:"quantity"`
			Note       string `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := d.medicalUsecase.TakeMedicine(r.Context(), body.MedicineID, body.Quantity, body.Note); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (d *Delivery) handleBillings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var body struct {
			PatientID   string `json:"patient_id"`
			EncounterID string `json:"encounter_id"`
			Amount      int    `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		b, err := d.billingUsecase.CreateBilling(r.Context(), body.PatientID, body.EncounterID, body.Amount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(b)

	case http.MethodGet:
		encounterID := r.URL.Query().Get("encounter_id")
		billings, err := d.billingUsecase.GetBillingsByEncounter(r.Context(), encounterID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(billings); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
		}
	}
}

func (d *Delivery) handleBillingByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/billings/"):]
	switch r.Method {
	case http.MethodGet:
		b, err := d.billingUsecase.GetBilling(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(b)

	case http.MethodPatch:
		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := d.billingUsecase.UpdateBillingStatus(r.Context(), id, body.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (d *Delivery) handlePayments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var body struct {
			BillingID string `json:"billing_id"`
			Amount    int    `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := d.billingUsecase.PayBilling(r.Context(), body.BillingID, body.Amount); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "paid"})

	case http.MethodGet:
		billingID := r.URL.Query().Get("billing_id")
		payments, err := d.billingUsecase.GetPaymentsByBilling(r.Context(), billingID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payments); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
		}
	}
}

func (d *Delivery) handlePaymentByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/payments/"):]
	switch r.Method {
	case http.MethodGet:
		payments, err := d.billingUsecase.GetPaymentsByBilling(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payments); err != nil {
			http.Error(w, "Encode error", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (d *Delivery) handleBPJSClaims(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var body struct {
			PatientID   string `json:"patient_id"`
			EncounterID string `json:"encounter_id"`
			Polyclinic  string `json:"polyclinic"`
			Amount      int    `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		sep, err := d.bpjsUsecase.CreateSEP(r.Context(), body.PatientID, body.EncounterID, body.Polyclinic, body.Amount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(sep)

	case http.MethodGet:
		noSEP := r.URL.Query().Get("no_sep")
		if noSEP == "" {
			http.Error(w, "no_sep query param required", http.StatusBadRequest)
			return
		}
		claim, err := d.bpjsUsecase.GetClaimStatus(r.Context(), noSEP)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claim)
	}
}

func (d *Delivery) handleBPJSByNoSEP(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/bpjs/claims/"):]
	switch r.Method {
	case http.MethodGet:
		claim, err := d.bpjsUsecase.GetClaimStatus(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claim)
	}
}
