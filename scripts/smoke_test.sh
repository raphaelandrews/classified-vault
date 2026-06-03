#!/usr/bin/env bash
set -e

BASE="http://localhost:8080"
PASS=0
FAIL=0

check() {
    local name="$1" status="$2" expected="$3"
    if [ "$status" = "$expected" ]; then
        echo "  OK     $name ($status)"
        PASS=$((PASS + 1))
    else
        echo "  FAIL   $name (got $status, expected $expected)"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== Classified Vault Smoke Tests ==="
echo ""

echo "[Health]"
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/health")
check "GET /health" "$status" "200"

echo ""
echo "[Auth]"
status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/auth/login" -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin123"}')
check "POST /auth/login (valid)" "$status" "200"
status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/auth/login" -H 'Content-Type: application/json' -d '{"username":"admin","password":"wrong"}')
check "POST /auth/login (invalid)" "$status" "401"

TOKEN=$(curl -s -X POST "$BASE/auth/login" -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

echo ""
echo "[Auth required]"
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/me")
check "GET /api/me (no token)" "$status" "401"
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/me" -H "Authorization: Bearer $TOKEN")
check "GET /api/me (with token)" "$status" "200"

echo ""
echo "[Documents]"
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/documents" -H "Authorization: Bearer $TOKEN")
check "GET /api/documents (empty list)" "$status" "200"

echo ""
echo "[Admin endpoints]"
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/users" -H "Authorization: Bearer $TOKEN")
check "GET /api/users" "$status" "200"
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/audit" -H "Authorization: Bearer $TOKEN")
check "GET /api/audit" "$status" "200"

echo ""
echo "[API docs]"
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/docs/index.html")
check "GET /docs/index.html" "$status" "200"

echo ""
echo "=== Results: $((PASS + FAIL)) tests, $PASS passed, $FAIL failed ==="
