# SIP System

In-memory SIP mutual-fund platform in Go with MVC-style controllers, service layer, repository abstractions, scheduler, and schedule strategies. This README reflects the current implementation in this repo.

## 1. Functional Requirements

### Core Requirements

1. Product Fetch and Selection
- Fetch all available mutual funds.
- Browse/search/filter funds.
- Select a fund and create an SIP on it.

2. SIP Duration and Setup
- Create an SIP by selecting:
  - Mutual fund
  - SIP amount
  - SIP mode: Weekly, Monthly, Quarterly
  - Start date
  - Optional step-up percentage
- Validate:
  - Fund exists and active
  - Amount > 0
  - Start date valid
  - Mode valid

3. SIP Execution
- Execute all due SIPs as per schedule:
  - Weekly: same weekday
  - Monthly: same day-of-month (fallback to last day if missing)
  - Quarterly: every 3 months (same day-of-month; fallback to last day)
- Each execution uses latest NAV at execution time (not creation time).
- Payment is external:
  - Execution triggers payment request
  - Callback marks success/failure
- Execution history is stored as installments.

4. Fund Cart / Portfolio View
- View SIPs grouped by state: Active, Paused, Stopped.
- Pause / unpause any SIP.
- Inspect SIP details + installment history.

### Bonus Requirements

5. Step-Up SIP
- Optional step-up percentage applied per installment.
- Example: base 1000, step-up 10%:
  - 1st: 1000
  - 2nd: 1100
  - 3rd: 1210

6. Lump Sum Payment for Missed Installments
- Pay multiple missed installments together.
- After successful lump sum, SIP can be reinstated.

## 2. Assumptions

- Single currency.
- Payment service is external and asynchronous.
- NAV fetched at execution time from pricing service.
- One fund per SIP.
- One user can have multiple SIPs.
- SIP states: Active, Paused, Stopped. Stopped is terminal.
- Pause skips future executions until unpaused.
- Scheduler runs periodically and checks due SIPs.
- Timezone uses `Asia/Kolkata` (configured in `cmd/main.go`).
- Repositories are interfaces with in-memory implementations for easy DB swap later.

## 3. High-Level Design

Layered design:
- Controller: HTTP parsing, response formatting
- Service: business rules
- Repository: storage abstractions
- Model: domain entities
- Scheduler: periodic due-SIP execution
- Strategy: frequency computation
- External integration: pricing + payment services

## 4. Folder Structure

```
sip-system/
├── cmd/
│   └── main.go
├── internal/
│   ├── controller/
│   │   ├── fund_controller.go
│   │   ├── sip_controller.go
│   │   └── portfolio_controller.go
│   ├── service/
│   │   ├── fund_service.go
│   │   ├── sip_service.go
│   │   ├── execution_service.go
│   │   ├── portfolio_service.go
│   │   ├── pricing_service.go
│   │   └── payment_service.go
│   ├── repository/
│   │   ├── fund_repository.go
│   │   ├── sip_repository.go
│   │   ├── installment_repository.go
│   │   └── user_repository.go
│   ├── model/
│   │   ├── fund.go
│   │   ├── sip.go
│   │   ├── installment.go
│   │   ├── enums.go
│   │   └── user.go
│   ├── scheduler/
│   │   └── sip_scheduler.go
│   ├── dto/
│   │   ├── create_sip_request.go
│   │   ├── lump_sum_request.go
│   │   └── payment_callback_request.go
│   ├── strategy/
│   │   ├── schedule_strategy.go
│   │   ├── weekly_strategy.go
│   │   ├── monthly_strategy.go
│   │   └── quarterly_strategy.go
│   └── util/
│       ├── validator.go
│       ├── id_generator.go
│       ├── time_util.go
│       └── http_util.go
└── go.mod
```

## 5. Core Domain Models (actual)

Mutual Fund
```go
type Fund struct {
    FundID   string
    Name     string
    AMC      string
    Category string
    RiskTag  string
    IsActive bool
}
```

User
```go
type User struct {
    UserID string
    Name   string
}
```

SIP (fixed-point amounts in paise)
```go
type SIP struct {
    SIPID  string
    UserID string
    FundID string

    Mode      SIPMode
    StartAt   time.Time
    NextRunAt time.Time

    AnchorDayOfMonth int
    AnchorWeekday    time.Weekday

    BaseAmountPaise int64
    StepUpEnabled   bool
    StepUpBps       int32

    Status    SIPStatus
    CreatedAt time.Time
    UpdatedAt time.Time
    Version   int64
}
```

Installment (NAV + units in micro-units)
```go
type Installment struct {
    InstallmentID    string
    SIPID            string
    SequenceNo       int64
    ScheduledAt      time.Time
    ExecutedAt       time.Time
    AmountPaise      int64
    NAVMic           int64
    UnitsMic         int64
    PaymentRequestID string
    PaymentStatus    PaymentStatus
    FailureReason    string
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

## 6. Enums (actual)

```go
type SIPMode string
const (
    SIPModeWeekly    SIPMode = "WEEKLY"
    SIPModeMonthly   SIPMode = "MONTHLY"
    SIPModeQuarterly SIPMode = "QUARTERLY"
)

type SIPStatus string
const (
    SIPStatusActive  SIPStatus = "ACTIVE"
    SIPStatusPaused  SIPStatus = "PAUSED"
    SIPStatusStopped SIPStatus = "STOPPED"
)

type PaymentStatus string
const (
    PaymentStatusPending PaymentStatus = "PENDING"
    PaymentStatusSuccess PaymentStatus = "SUCCESS"
    PaymentStatusFailed  PaymentStatus = "FAILED"
)
```

## 7. APIs (actual)

Fund APIs
- `GET /api/v1/fund/funds?query=&category=&amc=&riskTag=`
- `GET /api/v1/fund/funds?withPrice=true`
- `POST /api/v1/fund/funds`

SIP APIs
- `POST /api/v1/sip/sips`
- `GET /api/v1/sip/sips/{id}`
- `PATCH /api/v1/sip/sips/{id}/pause`
- `PATCH /api/v1/sip/sips/{id}/unpause`
- `PATCH /api/v1/sip/sips/{id}/stop`
- `POST /api/v1/sip/sips/{id}/catchup`

Portfolio APIs
- `GET /api/v1/sip/portfolio`

User APIs
- `POST /api/v1/user/users`
- `GET /api/v1/user/users`

Execution / Payment APIs
- `POST /api/v1/sip/payments/callback`

`userId` can be sent as query param (`?userId=...`) or request header `X-User-Id`.

## 8. In-Memory Repository Design (actual)

- Funds: `map[string]Fund`
- SIPs: `map[string]SIP`
- Installments: `map[string]Installment` + index `map[sipID][]installmentID`
- Users: `map[string]User`

All repos use `sync.RWMutex` and singleton constructors.

## 9. Service Responsibilities (actual)

FundService
- List and search funds

SIPService
- Create SIP
- Pause / Unpause / Stop
- Validate SIP setup

ExecutionService
- Find due SIPs
- Create installments
- Fetch latest NAV
- Trigger payment request
- Handle payment callback
- Advance next execution date
- Catch-up lump sum

PortfolioService
- List SIPs by user and state
- Provide SIP detail + installments

PricingService
- In-memory latest NAV source

PaymentService
- In-memory mock (idempotent by paymentRequestId)

## 10. Schedule Strategy Pattern

Interface:
```go
type ScheduleStrategy interface {
    NextRun(prev time.Time, sip model.SIP, loc *time.Location) time.Time
}
```

Implementations:
- Weekly
- Monthly
- Quarterly

This isolates cadence logic and makes it easy to add new frequencies.

## 11. SIP Creation Flow (actual)

- Validate user exists
- Validate fund exists and active
- Validate amount, mode, start date
- Create SIP with status `ACTIVE`
- Set `NextRunAt = StartAt`
- Persist

## 12. SIP Execution Flow (actual)

Scheduler loop:
- Every 5 seconds, find `ACTIVE` SIPs where `NextRunAt <= now`

Execution:
- Fetch latest NAV
- Compute amount (step-up if enabled)
- Create installment with `PENDING`
- Initiate payment request
- Advance `NextRunAt` using schedule strategy

Callback:
- On success, compute units and mark payment `SUCCESS`
- On failure, mark payment `FAILED`

## 13. Step-Up Calculation (actual)

Amount uses integer compounding by basis points:

```
amount = base * (1 + stepUpBps/10000)^(n-1)
```

Implemented in `internal/service/execution_service.go` with fixed-point integer math.

## 14. Lump Sum Flow (actual)

- User requests catch-up for `N` installments
- System computes total due using current NAV
- Creates single installment record with `PENDING`
- On success, SIP continues as normal

## 15. Validation Rules (actual)

- `fundId` valid and active
- `amount > 0`
- `startAt` non-zero
- `mode` in `{WEEKLY, MONTHLY, QUARTERLY}`
- `stepUpBps` in `[0, 10000]`
- State transitions:
  - `ACTIVE -> PAUSED`
  - `PAUSED -> ACTIVE`
  - `ACTIVE/PAUSED -> STOPPED`
  - `STOPPED -> ACTIVE` not allowed

## 16. Concurrency Notes

Repositories use `sync.RWMutex`. The scheduler scans due SIPs; there is no per-SIP lock or execution-in-progress flag. For higher correctness under concurrency, add per-SIP execution lock or optimistic state checks during `Update`.

## 17. Testability (actual)

- Integration test: `internal/controller/api_integration_test.go` (exercises all endpoints with `httptest`)
- Repo and service tests can be added around schedule edges, step-up, and failure handling.

## 18. Optimizations / Improvements

- Replace in-memory repos with DB (Postgres/MySQL)
- Add historical NAV store for accurate back-dated catch-up
- Add per-SIP execution locks to avoid double execution
- Add payment retry policies and audit logs
- Use priority queue for `NextRunAt` for scalable scheduling
- Add idempotency keys on API requests

## Run

```bash
go run ./cmd
```

## Quickstart (Sequential)

1. Start the server:

```bash
go run ./cmd
```

2. Create users:

```bash
curl -s -X POST http://localhost:8080/api/v1/user/users \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","name":"User One"}'

curl -s -X POST http://localhost:8080/api/v1/user/users \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-2","name":"User Two"}'
```

3. Create funds with NAV:

```bash
curl -s -X POST http://localhost:8080/api/v1/fund/funds \
  -H 'Content-Type: application/json' \
  -d '{"fundId":"fund-1","name":"Bluechip Equity Growth","amc":"Alpha AMC","category":"Equity","riskTag":"High","isActive":true,"navMic":12543210}'

curl -s -X POST http://localhost:8080/api/v1/fund/funds \
  -H 'Content-Type: application/json' \
  -d '{"fundId":"fund-2","name":"Balanced Advantage Plan","amc":"Zen AMC","category":"Hybrid","riskTag":"Moderate","isActive":true,"navMic":3022111}'

curl -s -X POST http://localhost:8080/api/v1/fund/funds \
  -H 'Content-Type: application/json' \
  -d '{"fundId":"fund-3","name":"Government Bond Income","amc":"Safe AMC","category":"Debt","riskTag":"Low","isActive":true,"navMic":1810050}'
```

4. List funds and prices:

```bash
curl -s http://localhost:8080/api/v1/fund/funds
curl -s http://localhost:8080/api/v1/fund/funds?withPrice=true
```

5. Create SIPs:

```bash
curl -s -X POST http://localhost:8080/api/v1/sip/sips \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","fundId":"fund-1","mode":"WEEKLY","startAt":"2026-03-06T09:30:00+05:30","baseAmountPaise":100000,"stepUpEnabled":false,"stepUpBps":0}'

curl -s -X POST http://localhost:8080/api/v1/sip/sips \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","fundId":"fund-2","mode":"MONTHLY","startAt":"2026-03-06T09:30:00+05:30","baseAmountPaise":200000,"stepUpEnabled":true,"stepUpBps":500}'

curl -s -X POST http://localhost:8080/api/v1/sip/sips \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","fundId":"fund-3","mode":"QUARTERLY","startAt":"2026-03-06T09:30:00+05:30","baseAmountPaise":150000,"stepUpEnabled":true,"stepUpBps":1000}'
```

6. Portfolio:

```bash
curl -s http://localhost:8080/api/v1/sip/portfolio?userId=user-1
```

7. Run the full automated flow:

```bash
scripts/test_full_flow.sh
```

## Curl Test Commands

Create user:

```bash
curl -s -X POST http://localhost:8080/api/v1/user/users \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","name":"Demo User"}'
```

List users:

```bash
curl -s http://localhost:8080/api/v1/user/users
```

Create fund (with NAV):

```bash
curl -s -X POST http://localhost:8080/api/v1/fund/funds \
  -H 'Content-Type: application/json' \
  -d '{"fundId":"fund-1","name":"Bluechip Equity Growth","amc":"Alpha AMC","category":"Equity","riskTag":"High","isActive":true,"navMic":12543210}'
```

List funds:

```bash
curl -s http://localhost:8080/api/v1/fund/funds
```

List funds with latest NAV:

```bash
curl -s http://localhost:8080/api/v1/fund/funds?withPrice=true
```

Create SIP:

```bash
curl -s -X POST http://localhost:8080/api/v1/sip/sips \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","fundId":"fund-1","mode":"MONTHLY","startAt":"2026-03-05T09:30:00+05:30","baseAmountPaise":100000,"stepUpEnabled":true,"stepUpBps":1000}'
```

Get portfolio:

```bash
curl -s http://localhost:8080/api/v1/sip/portfolio?userId=user-1
```

Get SIP details:

```bash
curl -s http://localhost:8080/api/v1/sip/sips/{sipId}?userId=user-1
```

Pause SIP:

```bash
curl -s -X PATCH http://localhost:8080/api/v1/sip/sips/{sipId}/pause?userId=user-1
```

Unpause SIP:

```bash
curl -s -X PATCH http://localhost:8080/api/v1/sip/sips/{sipId}/unpause?userId=user-1
```

Stop SIP:

```bash
curl -s -X PATCH http://localhost:8080/api/v1/sip/sips/{sipId}/stop?userId=user-1
```

Catch-up (lump sum):

```bash
curl -s -X POST http://localhost:8080/api/v1/sip/sips/{sipId}/catchup?userId=user-1 \
  -H 'Content-Type: application/json' \
  -d '{"numInstallments":2}'
```

Payment callback:

```bash
curl -s -X POST http://localhost:8080/api/v1/sip/payments/callback \
  -H 'Content-Type: application/json' \
  -d '{"paymentRequestId":"{paymentRequestId}","status":"SUCCESS","failureReason":""}'
```

## Test

```bash
GOCACHE=/tmp/go-build go test ./...
```

