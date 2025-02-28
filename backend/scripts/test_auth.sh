#!/bin/bash

# Base URL
BASE_URL="http://localhost:8080"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Testing Authentication System${NC}\n"

# 1. Register a new user
echo "1. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/auth/register \
-H "Content-Type: application/json" \
-d '{
"email": "test@example.com",
"password": "password123"
}')
echo "Response: $REGISTER_RESPONSE"
echo -e "\n"

# 2. Login
echo "2. Testing login..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/auth/login \
-H "Content-Type: application/json" \
-d '{
"email": "test@example.com",
"password": "password123"
}')
echo "Response: $LOGIN_RESPONSE"

# Extract tokens from login response
ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | sed 's/.*"access_token":"\([^"]*\).*/\1/')
REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | sed 's/.*"refresh_token":"\([^"]*\).*/\1/')
echo -e "\n"

# 3. Test protected endpoint
echo "3. Testing protected endpoint (/api/me)..."
ME_RESPONSE=$(curl -s $BASE_URL/api/me \
-H "Authorization: Bearer $ACCESS_TOKEN")
echo "Response: $ME_RESPONSE"
echo -e "\n"

# 4. Test token refresh
echo "4. Testing token refresh..."
REFRESH_RESPONSE=$(curl -s -X POST $BASE_URL/auth/refresh \
-H "Content-Type: application/json" \
-d "{
\"refresh_token\": \"$REFRESH_TOKEN\"
}")
echo "Response: $REFRESH_RESPONSE"
echo -e "\n"

# 5. Test invalid credentials
echo "5. Testing invalid login..."
INVALID_LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/auth/login \
-H "Content-Type: application/json" \
-d '{
"email": "test@example.com",
"password": "wrongpassword"
}')
echo "Response: $INVALID_LOGIN_RESPONSE"
echo -e "\n"

echo -e "${GREEN}Testing complete!${NC}"
