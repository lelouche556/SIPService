#!/usr/bin/env bash
set -euo pipefail

BASE="http://localhost:8080"

http_code() {
  curl -s -o /tmp/resp.json -w "%{http_code}" "$@"
}

assert_code() {
  local code="$1"
  local expected="$2"
  if [ "$code" != "$expected" ]; then
    echo "Expected HTTP $expected, got $code" >&2
    cat /tmp/resp.json >&2 || true
    exit 1
  fi
}

assert_code_any() {
  local code="$1"; shift
  for expected in "$@"; do
    if [ "$code" = "$expected" ]; then
      return 0
    fi
  done
  echo "Expected HTTP one of [$*], got $code" >&2
  cat /tmp/resp.json >&2 || true
  exit 1
}

get_field() {
  local key="$1"
  if command -v jq >/dev/null 2>&1; then
    jq -r ".${key}" /tmp/resp.json
  else
    sed -n "s/.*\"${key}\":\"\([^\"]*\)\".*/\1/p" /tmp/resp.json
  fi
}

echo "Creating users..."
code=$(http_code -X POST "$BASE/api/v1/user/users" \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","name":"User One"}')
assert_code_any "$code" 201 409

code=$(http_code -X POST "$BASE/api/v1/user/users" \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-2","name":"User Two"}')
assert_code_any "$code" 201 409

# Duplicate user
code=$(http_code -X POST "$BASE/api/v1/user/users" \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","name":"User One"}')
assert_code "$code" 409

echo "Listing users..."
code=$(http_code "$BASE/api/v1/user/users")
assert_code "$code" 200

# Validation: missing name
code=$(http_code -X POST "$BASE/api/v1/user/users" \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-3"}')
assert_code "$code" 400


echo "Creating funds..."
code=$(http_code -X POST "$BASE/api/v1/fund/funds" \
  -H 'Content-Type: application/json' \
  -d '{"fundId":"fund-1","name":"Bluechip Equity Growth","amc":"Alpha AMC","category":"Equity","riskTag":"High","isActive":true,"navMic":12543210}')
assert_code_any "$code" 201 409

code=$(http_code -X POST "$BASE/api/v1/fund/funds" \
  -H 'Content-Type: application/json' \
  -d '{"fundId":"fund-2","name":"Balanced Advantage Plan","amc":"Zen AMC","category":"Hybrid","riskTag":"Moderate","isActive":true,"navMic":3022111}')
assert_code_any "$code" 201 409

code=$(http_code -X POST "$BASE/api/v1/fund/funds" \
  -H 'Content-Type: application/json' \
  -d '{"fundId":"fund-3","name":"Government Bond Income","amc":"Safe AMC","category":"Debt","riskTag":"Low","isActive":true,"navMic":1810050}')
assert_code_any "$code" 201 409

# Duplicate fund
code=$(http_code -X POST "$BASE/api/v1/fund/funds" \
  -H 'Content-Type: application/json' \
  -d '{"fundId":"fund-1","name":"Duplicate","amc":"Alpha AMC","category":"Equity","riskTag":"High","isActive":true,"navMic":12543210}')
assert_code "$code" 409

# Validation: missing fields
code=$(http_code -X POST "$BASE/api/v1/fund/funds" \
  -H 'Content-Type: application/json' \
  -d '{"fundId":"fund-4","name":"","amc":"","category":"","riskTag":"","navMic":0}')
assert_code "$code" 400

# List funds with price
code=$(http_code "$BASE/api/v1/fund/funds?withPrice=true")
assert_code "$code" 200


echo "Creating SIPs for user-1..."
now=$(date -u -v+1M +%Y-%m-%dT%H:%M:%SZ)

code=$(http_code -X POST "$BASE/api/v1/sip/sips" \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","fundId":"fund-1","mode":"WEEKLY","startAt":"'$now'","baseAmountPaise":100000,"stepUpEnabled":false,"stepUpBps":0}')
assert_code "$code" 201
sip_weekly_id=$(get_field "sipId")

code=$(http_code -X POST "$BASE/api/v1/sip/sips" \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","fundId":"fund-2","mode":"MONTHLY","startAt":"'$now'","baseAmountPaise":200000,"stepUpEnabled":true,"stepUpBps":500}')
assert_code "$code" 201
sip_monthly_id=$(get_field "sipId")

code=$(http_code -X POST "$BASE/api/v1/sip/sips" \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","fundId":"fund-3","mode":"QUARTERLY","startAt":"'$now'","baseAmountPaise":150000,"stepUpEnabled":true,"stepUpBps":1000}')
assert_code "$code" 201
sip_quarterly_id=$(get_field "sipId")

# Validation: invalid fund
code=$(http_code -X POST "$BASE/api/v1/sip/sips" \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","fundId":"fund-unknown","mode":"MONTHLY","startAt":"'$now'","baseAmountPaise":100000}')
assert_code "$code" 404

# Validation: invalid amount
code=$(http_code -X POST "$BASE/api/v1/sip/sips" \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","fundId":"fund-1","mode":"MONTHLY","startAt":"'$now'","baseAmountPaise":0}')
assert_code "$code" 400

# Validation: invalid mode
code=$(http_code -X POST "$BASE/api/v1/sip/sips" \
  -H 'Content-Type: application/json' \
  -d '{"userId":"user-1","fundId":"fund-1","mode":"YEARLY","startAt":"'$now'","baseAmountPaise":100000}')
assert_code "$code" 400

# Pause/unpause/stop
code=$(http_code -X PATCH "$BASE/api/v1/sip/sips/$sip_weekly_id/pause?userId=user-1")
assert_code "$code" 200

code=$(http_code -X PATCH "$BASE/api/v1/sip/sips/$sip_weekly_id/unpause?userId=user-1")
assert_code "$code" 200

code=$(http_code -X PATCH "$BASE/api/v1/sip/sips/$sip_quarterly_id/stop?userId=user-1")
assert_code "$code" 200

# Stop -> unpause should fail
code=$(http_code -X PATCH "$BASE/api/v1/sip/sips/$sip_quarterly_id/unpause?userId=user-1")
assert_code "$code" 400

# Catch-up creates installment
code=$(http_code -X POST "$BASE/api/v1/sip/sips/$sip_monthly_id/catchup?userId=user-1" \
  -H 'Content-Type: application/json' \
  -d '{"numInstallments":2}')
assert_code "$code" 200
payReq=$(get_field "paymentRequestId")

# Payment callback success
code=$(http_code -X POST "$BASE/api/v1/sip/payments/callback" \
  -H 'Content-Type: application/json' \
  -d '{"paymentRequestId":"'$payReq'","status":"SUCCESS","failureReason":""}')
assert_code "$code" 200

# Idempotent callback
code=$(http_code -X POST "$BASE/api/v1/sip/payments/callback" \
  -H 'Content-Type: application/json' \
  -d '{"paymentRequestId":"'$payReq'","status":"SUCCESS","failureReason":""}')
assert_code "$code" 200

# Invalid callback
code=$(http_code -X POST "$BASE/api/v1/sip/payments/callback" \
  -H 'Content-Type: application/json' \
  -d '{"paymentRequestId":"does-not-exist","status":"SUCCESS","failureReason":""}')
assert_code "$code" 404

# Get portfolio and SIP details
code=$(http_code "$BASE/api/v1/sip/portfolio?userId=user-1")
assert_code "$code" 200

code=$(http_code "$BASE/api/v1/sip/sips/$sip_weekly_id?userId=user-1")
assert_code "$code" 200

# Concurrency smoke test: hit pause endpoint in parallel
for i in {1..5}; do
  curl -s -X PATCH "$BASE/api/v1/sip/sips/$sip_weekly_id/pause?userId=user-1" >/dev/null &
done
wait

code=$(http_code "$BASE/api/v1/sip/sips/$sip_weekly_id?userId=user-1")
assert_code "$code" 200

# Concurrency smoke test: execute catchup in parallel
for i in {1..3}; do
  curl -s -X POST "$BASE/api/v1/sip/sips/$sip_monthly_id/catchup?userId=user-1" \
    -H 'Content-Type: application/json' \
    -d '{"numInstallments":1}' >/dev/null &
done
wait

echo "All checks passed."
