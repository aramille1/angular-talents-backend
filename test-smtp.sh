#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Angular Talents SMTP Configuration Test${NC}"
echo "This script will check your SMTP configuration and test sending an email."
echo

# Load environment variables from .env file
if [ -f .env ]; then
  echo -e "${GREEN}Loading configuration from .env file...${NC}"
  export $(grep -v '^#' .env | xargs)
else
  echo -e "${RED}No .env file found. Using environment variables only.${NC}"
fi

# Get SMTP configuration
SMTP_HOST=${SMTP_HOST:-"live.smtp.mailtrap.io"}
SMTP_PORT=${SMTP_PORT:-"587"}
SMTP_USER=${SMTP_USER:-"api"}
SMTP_PASSWORD=${SMTP_PASSWORD:-$MAILTRAP_TOKEN}

# Print current configuration (masking password)
echo
echo -e "${YELLOW}Current SMTP Configuration:${NC}"
echo -e "SMTP Host: ${SMTP_HOST}"
echo -e "SMTP Port: ${SMTP_PORT}"
echo -e "SMTP User: ${SMTP_USER}"
if [ -n "$SMTP_PASSWORD" ]; then
  VISIBLE_PART="${SMTP_PASSWORD:0:4}...${SMTP_PASSWORD: -4}"
  echo -e "SMTP Password: ${VISIBLE_PART} (${#SMTP_PASSWORD} characters)"
else
  echo -e "${RED}SMTP Password: Not set${NC}"
fi

# Prompt for test email address
echo
echo -e "${YELLOW}Please enter a test recipient email address:${NC}"
read -p "Email: " TEST_EMAIL

if [ -z "$TEST_EMAIL" ]; then
  echo -e "${RED}No email provided. Using test@example.com${NC}"
  TEST_EMAIL="test@example.com"
fi

# Create a temporary file with the email content
TEMP_EMAIL=$(mktemp)
cat > $TEMP_EMAIL << EOF
From: Angular Talents <hello@angular-talents.mailtrap.io>
To: $TEST_EMAIL
Subject: SMTP Test Email
MIME-Version: 1.0
Content-Type: text/html; charset="UTF-8"

<!DOCTYPE html>
<html>
<head>
    <title>SMTP Test</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; border-radius: 5px; padding: 20px; margin-bottom: 20px;">
        <h2 style="color: #333;">SMTP Configuration Test</h2>
        <p>This is a test email sent from the Angular Talents backend.</p>
        <p>If you received this email, your SMTP configuration is working correctly!</p>
        <hr>
        <p><strong>Configuration:</strong></p>
        <ul>
            <li>Host: $SMTP_HOST</li>
            <li>Port: $SMTP_PORT</li>
            <li>User: $SMTP_USER</li>
            <li>Password: ${SMTP_PASSWORD:0:2}...${SMTP_PASSWORD: -2}</li>
        </ul>
        <p>Best regards,<br>The Angular Talents Team</p>
    </div>
</body>
</html>
EOF

echo
echo -e "${BLUE}Sending test email to $TEST_EMAIL...${NC}"

# Check if curl is available
if ! command -v curl &> /dev/null; then
  echo -e "${RED}curl command not found. Please install curl and try again.${NC}"
  rm $TEMP_EMAIL
  exit 1
fi

# Check if openssl is available
if ! command -v openssl &> /dev/null; then
  echo -e "${RED}openssl command not found. Please install openssl and try again.${NC}"
  rm $TEMP_EMAIL
  exit 1
fi

# Use OpenSSL for SMTP with STARTTLS
(
  echo "EHLO localhost"
  sleep 1
  echo "STARTTLS"
  sleep 1
  openssl s_client -starttls smtp -connect $SMTP_HOST:$SMTP_PORT -quiet 2>/dev/null <<EOF
EHLO localhost
AUTH LOGIN
$(echo -n $SMTP_USER | base64)
$(echo -n $SMTP_PASSWORD | base64)
MAIL FROM: <hello@angular-talents.mailtrap.io>
RCPT TO: <$TEST_EMAIL>
DATA
$(cat $TEMP_EMAIL)
.
QUIT
EOF
) | tee /tmp/smtp_output.log

# Check for errors in the output
if grep -i "Authentication failed" /tmp/smtp_output.log > /dev/null; then
  echo -e "${RED}Authentication failed. Check your username and password.${NC}"
elif grep -i "250 OK" /tmp/smtp_output.log > /dev/null; then
  echo -e "${GREEN}Email sent successfully!${NC}"
else
  echo -e "${YELLOW}Email sending may have failed. Check the output above for errors.${NC}"
fi

# Clean up
rm $TEMP_EMAIL
rm /tmp/smtp_output.log

echo
echo -e "${BLUE}Test completed. For more detailed testing, try:${NC}"
echo "1. Checking your Mailtrap inbox for the test email"
echo "2. Running the Angular Talents backend and testing the signup or verification flow"

echo
echo -e "${YELLOW}SMTP Configuration summary:${NC}"
echo -e "To use SMTP in your Angular Talents backend, ensure your .env file contains:"
echo "SMTP_HOST=$SMTP_HOST"
echo "SMTP_PORT=$SMTP_PORT"
echo "SMTP_USER=$SMTP_USER"
echo "SMTP_PASSWORD=[your_password]"
echo
echo -e "${GREEN}You can also still use the API token as a fallback:${NC}"
echo "MAILTRAP_TOKEN=[your_token]"
echo
