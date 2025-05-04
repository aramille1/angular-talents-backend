<!-- PROJECT LOGO -->
<br />
<p align="center">
  <h3 align="center">Reverse Job Board Backend</h3>

  <p align="center">
    Find out the best talents out there
    <br />
    <a href="https://www.notion.so/axelmtn/La-Perette-b9cb65b6f7e34df7abc43d80412428c4"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://laperette-client.herokuapp.com/">View Demo</a>
    ·
    <a href="https://github.com/paradoux/reverse-job-board-backend/issues">Report Bug</a>
    ·
    <a href="https://github.com/paradoux/reverse-job-board-backend/issues">Request Feature</a>
  </p>
</p>

<!-- TABLE OF CONTENTS -->

## Table of Contents

- [About the Project](#about-the-project)
  - [Built With](#built-with)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  <!-- - [Usage](#usage) -->
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)

<!-- ABOUT THE PROJECT -->

## About The Project

### Documentation

Just head to our [Notion page](https://www.notion.so/axelmtn/Reverse-Job-Board-Project-75efc9e3409a42f592690f3807f7154e?pvs=4) to understand how to use our platform and have a look at how it is architectured.

### Built With

- [Go](https://go.dev/)
- [MongoDB](https://www.mongodb.com/)

<!-- GETTING STARTED -->

## Getting Started

To get a local copy up and running follow these simple steps.

### Prerequisites

- Install Go
- Install MongoDB

### Installation

1. Clone the repo

```sh
git clone https://github.com/paradoux/reverse-job-board-backend
```

2. Launch the local database accessible at `"mongodb://127.0.0.1:27017"`

```sh
brew services start mongodb-community@6.0
```

3. Launch the server

```sh
go run .
```

## Email Configuration

### SMTP Configuration (Recommended)
The application now supports email sending via SMTP, which is more reliable than the API method.

Add the following to your `.env` file:
```
SMTP_HOST=live.smtp.mailtrap.io
SMTP_PORT=587
SMTP_USER=your_mailtrap_username
SMTP_PASSWORD=your_mailtrap_password
```

For Mailtrap specifically:
1. Log in to your Mailtrap account
2. Go to Email Testing > Inboxes > Select your inbox
3. Click on "SMTP Settings"
4. Copy the credentials for the "Nodemailer" integration

### Testing SMTP Configuration
A test script is included to verify your SMTP configuration:

```bash
./test-smtp.sh
```

This script will:
1. Load your SMTP configuration from `.env`
2. Prompt for a test email address
3. Send a test email using your SMTP settings
4. Provide feedback on the success or failure of the email sending

### Legacy API Configuration
The application still supports the Mailtrap API method as a fallback. If you prefer to use this method, set the following in your `.env` file:

```
MAILTRAP_TOKEN=your_mailtrap_api_token
VERIFICATION_TEMPLATE_ID=your_mailtrap_template_id
RECRUITER_APPROVAL_TEMPLATE_ID=your_mailtrap_recruiter_approval_template_id
```

## Running the Application

### With Docker
```bash
docker build -t angular-talents-backend .
docker run -p 8080:8080 -e SMTP_PASSWORD=your_password angular-talents-backend
```

### Without Docker
```bash
go run main.go
```

## Environment Variables
See `.env.example` for all available configuration options.

<!-- ROADMAP -->

## Roadmap

See the [open issues](https://github.com/paradoux/reverse-job-board-backend/issues) for a list of proposed features (and known issues).

<!-- CONTRIBUTING -->

## Contributing

Any contributions you make are **greatly appreciated**, especially styling!

1. Fork the Project
2. Create your Feature Branch using your initials (`git checkout -b am/amazing-feature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin am/amazing-feature`)
5. Open a Pull Request

<!-- LICENSE -->

## License

Distributed under the MIT License. See `LICENSE` for more information.

<!-- CONTACT -->

## Contact

Axel Martin - mtn.axel@gmail.com

[Github](https://github.com/paradoux) - [LinkedIn](https://www.linkedin.com/in/martinaxel/)

Ramil Assanov - aramille@gmail.com

[Github](https://github.com/aramille1) - [LinkedIn](https://de.linkedin.com/in/ramil-assanov-31194940)
