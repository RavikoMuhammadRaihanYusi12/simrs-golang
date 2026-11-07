# SIMRS Backend - Phase 3 Status

## Overview
Phase 3 (Billing + BPJS + SATUSEHAT) is **mostly complete but not fully finished**. Core functionality is implemented and running, but there are still several gaps for production readiness.

## ✅ Implemented (Phase 3 Core)

### 1. Billing Module
- **Models**: Billing, Invoice, Payment
- **Repository**: SQLiteBillingRepository with full CRUD
- **Usecase**: BillingUsecase (CreateBilling, PayBilling, financial report)
- **API endpoints**:
  - `POST /api/v1/billings` - Create billing
  - `GET /api/v1/billings/:id` - Get billing by ID
  - `PATCH /api/v1/billings/:id` - Update billing status
  - `POST /api/v1/payments` - Record payment
  - `GET /api/v1/payments/billing_id` - List payments by billing

### 2. BPJS Module
- **Models**: BPJSSEP, BPJSClaim
- **Usecase**: BPJSUsecase (CreateSEP, GetClaimStatus)
- **API endpoints**:
  - `POST /api/v1/bpjs/claims` - SEP generation
  - `GET /api/v1/bpjs/claims/:no_sep` - Claim status lookup

### 3. SATUSEHAT Module
- **Model**: SATUSEHATPayload struct ready for async integration
- **Planned**: SATUSEHAT API endpoint for async medical record integration

### 4. Core Infrastructure
- **Server**: Runs on `:8080`, builds and vets
- **Clean Architecture**: Handler → Usecase → Repository layers
- **JWT auth**: HS256 with role-based access control

## ⚠️ Still Incomplete (Next Move Items)

### 1. Database Migration
- **Status**: Still using SQLite in-memory
- **Goal**: Migrate to PostgreSQL with connection pool
- **Blockers**: Production DB setup needed

### 2. Redis Integration
- **Status**: Real-time queue still in-memory only
- **Goal**: Redis Pub/Sub for queuing

### 3. Full BPJS V-Claim Integration
- **Status**: SEP generation only
- **Missing**: HMAC-SHA256 signature, eligibility check against Kemenkes API
- **Blockers**: External API keys/certificates needed

### 4. Unit Tests
- **Status**: Started test infrastructure
- **Missing**: Comprehensive test coverage for billing/medical usecase

### 5. CI/CD Pipeline
- **Status**: Pipeline framework created in `.github/workflows/ci.yml`
- **Missing**: Actual Go/Next.js build and test steps configured

## Current State

### ✅ Running Features
```bash
go run ./cmd/server
# Server starts on :8080
# All Billing + BPJS API endpoints functional
# Build + vet passes
```

### ⚠️ Limitations
- SQLite only (not PostgreSQL)
- In-memory queue (not Redis)
- Basic BPJS integration (no V-Claim SOAP)
- Limited test coverage
- CI/CD not fully configured

### 📋 Test Status
- Unit test framework: Created billing_test.go (needs refinement)
- Integration tests: Not yet written
- Coverage: Minimal, primarily stubbed

## Roadmap for Completion

1. **Database Migration**: Set up PostgreSQL instance and update connection logic
2. **Redis Integration**: Implement Pub/Sub for queuing
3. **BPJS V-Claim**: Configure external API integration
4. **Complete Tests**: Write comprehensive unit tests
5. **CI/CD**: Configure actual build/test pipeline

## Files Modified/Created

### Backend
- `internal/domain/billing.go` - Added Payment, cleaned SATUSEHATPayload
- `internal/repository/billing.go` - Billing repository
- `internal/usecase/billing.go` - Billing usecase
- `internal/usecase/bpjs.go` - BPJS usecase
- `internal/delivery/handler.go` - All Billing + BPJS routes
- `cmd/server/main.go` - Wiring billing + BPJS
- `internal/usecase/billing_test.go` - Test infrastructure

### DevOps
- `.github/workflows/ci.yml` - CI pipeline framework
- `STATUS.md` - This status document

## Summary

Phase 3 provides **core billing and BPJS functionality** that works:
- ✅ Billing CRUD operations
- ✅ Payment processing
- ✅ BPJS SEP generation
- ✅ Claim status lookup
- ✅ SATUSEHAT payload ready

**Remaining gaps** are primarily infrastructure and integration items needed for production readiness:
- PostgreSQL migration
- Redis integration
- Full external BPJS integration
- Comprehensive testing
- CI/CD pipeline

**Recommendation**: The core Phase 3 features are production-ready for development/testing. Focus on completing the "Next Move" items for full production readiness.