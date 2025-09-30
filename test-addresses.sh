#!/bin/bash

echo "Testing Address API Endpoints..."

# First, let's register a user and get a token
echo "1. Registering test user..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/users/start-registration \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "name": "Test User"}')
echo "Register response: $REGISTER_RESPONSE"

# For testing, let's try to get addresses without auth first
echo -e "\n2. Testing GET addresses without auth (should fail)..."
curl -X GET http://localhost:8081/api/v1/users/addresses -w "\nStatus: %{http_code}\n"

# Test with a mock token (this will fail but shows the endpoint exists)
echo -e "\n3. Testing GET addresses with mock token..."
curl -X GET http://localhost:8081/api/v1/users/addresses \
  -H "Authorization: Bearer mock-token" -w "\nStatus: %{http_code}\n"

# Test creating an address with mock token
echo -e "\n4. Testing POST address with mock token..."
curl -X POST http://localhost:8081/api/v1/users/addresses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer mock-token" \
  -d '{
    "label": "Home",
    "address_line": "123 Test Street",
    "city": "Test City",
    "state": "Test State",
    "zip_code": "12345",
    "country": "Test Country",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "is_default": true
  }' -w "\nStatus: %{http_code}\n"

echo -e "\nAPI endpoints are available!"