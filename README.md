<!-- PROJECT LOGO -->
<br />
<p align="center">
  <h3 align="center">Angular Talents Backend</h3>

  <p align="center">
    Find out the best Angular talents out there
    <br />

    <a href="https://github.com/aramille1/angular-talents-backend/issues">Report Bug</a>
    ·
    <a href="https://github.com/aramille1/angular-talents-backend/issues">Request Feature</a>
  </p>
</p>

<!-- TABLE OF CONTENTS -->

## Table of Contents

- [About the Project](#about-the-project)
  - [Built With](#built-with)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
  <!-- - [Usage](#usage) -->
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)

<!-- ABOUT THE PROJECT -->

## About The Project

A talent discovery platform specifically for Angular developers that allows software engineers to create profiles and recruiters to browse through them. The platform aims to invert the traditional job search process.

### Built With

- [Go](https://go.dev/)
- [MongoDB Atlas](https://www.mongodb.com/atlas/database)

<!-- GETTING STARTED -->

## Getting Started

To get a local copy up and running follow these simple steps.

### Prerequisites

- Install Go
- MongoDB Atlas account (or a local MongoDB installation)

### Installation

1. Clone the repo

```sh
git clone https://github.com/aramille1/angular-talents-backend
```

2. Set up your MongoDB connection

Either:
- Use MongoDB Atlas (cloud)
- Or launch a local MongoDB instance at `"mongodb://127.0.0.1:27017"`

```sh
brew services start mongodb-community@6.0
```

3. Launch the server

```sh
go run .
```

### Configuration

#### Secure Database Connection

For security reasons, it's recommended to use environment variables for database credentials:

1. Create a `.env` file in the root directory (make sure it's in your .gitignore)
2. Add your MongoDB connection details:

```
MONGODB_USERNAME=your_username
MONGODB_PASSWORD=your_password
MONGODB_CLUSTER=your_cluster.mongodb.net
MONGODB_DATABASE=your_database_name
```

3. Update connection.go to use environment variables (example):

```go
// Load environment variables
username := os.Getenv("MONGODB_USERNAME")
password := os.Getenv("MONGODB_PASSWORD")
cluster := os.Getenv("MONGODB_CLUSTER")
dbName := os.Getenv("MONGODB_DATABASE")

uri := fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority",
    username,
    password,
    cluster)
```

<!-- ROADMAP -->

## Roadmap

See the [open issues](https://github.com/aramille1/angular-talents-backend/issues) for a list of proposed features (and known issues).

<!-- CONTRIBUTING -->

## Contributing

Any contributions you make are **greatly appreciated**, especially styling!

1. Fork the Project
2. Create your Feature Branch using your initials (`git checkout -b ra/amazing-feature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin ra/amazing-feature`)
5. Open a Pull Request

<!-- LICENSE -->

## License

Distributed under the MIT License. See `LICENSE` for more information.

<!-- CONTACT -->

## Contact

Ramil Assanov - aramille@gmail.com

[Github](https://github.com/aramille1) - [LinkedIn](https://de.linkedin.com/in/ramil-assanov-31194940)

*Initially created by Axel Martin, but later he dropped out and I duplicated the backend and continued by myself (Ramil).
