#!/bin/bash

# Test Mailtrap API script

echo "Testing Mailtrap API token..."
echo "-----------------------------"

# Load token from environment
if [ -f .env ]; then
  source .env
  echo "Loaded environment from .env file"
else
  echo "No .env file found, using environment variables"
fi

# Check if token is set
if [ -z "$MAILTRAP_TOKEN" ]; then
  echo "ERROR: MAILTRAP_TOKEN is not set"
  exit 1
fi

# Show token info (partially masked)
TOKEN_LENGTH=${#MAILTRAP_TOKEN}
TOKEN_SAMPLE="${MAILTRAP_TOKEN:0:4}...${MAILTRAP_TOKEN: -4}"
echo "Token length: $TOKEN_LENGTH"
echo "Token sample: $TOKEN_SAMPLE"

# Create test email JSON
RECIPIENT="test@example.com" # Replace with your test email if needed
JSON_DATA=$(cat << EOF
{
  "from": {
    "email": "hello.angulartalents@gmail.com",
    "name": "Mailtrap Test"
  },
  "to": [
    {
      "email": "$RECIPIENT"
    }
  ],
  "subject": "API Test - Angular Talents",
  "text": "This is a test email from Angular Talents to verify the Mailtrap API is working.",
  "html": "<h1>Test Email</h1><p>This is a test email to verify the Mailtrap API is working.</p>"
}
EOF
)

echo ""
echo "Sending test request to Mailtrap API..."

# Send request
RESPONSE=$(curl -s -X POST \
  https://send.api.mailtrap.io/api/send \
  -H "Authorization: Bearer $MAILTRAP_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$JSON_DATA")

echo ""
echo "Response from Mailtrap API:"
echo "$RESPONSE" | json_pp 2>/dev/null || echo "$RESPONSE"

# Check for success in response
if echo "$RESPONSE" | grep -q "success"; then
  echo ""
  echo "SUCCESS: Test email sent successfully!"
elif echo "$RESPONSE" | grep -q "Unauthorized"; then
  echo ""
  echo "ERROR: Unauthorized. Your token is invalid or expired."
  echo "Please check your Mailtrap API token (not SMTP credentials)."
  echo "Visit https://mailtrap.io/api and get the API token for the 'Email API'"
else
  echo ""
  echo "ERROR: Failed to send test email. See response for details."
fi

echo ""
echo "Test completed."
