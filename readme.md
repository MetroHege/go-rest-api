# Go REST API

This is a sample REST API built with Go, Fiber, and MongoDB. The API provides endpoints for managing animals, species, and categories.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [API Documentation](#api-documentation)
- [Project Structure](#project-structure)
- [Environment Variables](#environment-variables)
- [Contributing](#contributing)
- [License](#license)

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/MetroHege/go-rest-api.git
   cd go-rest-api
   ```

2. Install dependencies:

   ```sh
   go mod tidy
   ```

3. Set up your environment variables. Create a `.env` file in the root directory and add the following variables:

   ```env
   MONGODB_URI=mongodb://localhost:27017
   PORT=5000
   ```

4. Generate Swagger documentation:

   ```sh
   swag init
   ```

## Usage

1. Run the application:

   ```sh
   go run main.go
   ```

2. The server will start on the port specified in the `.env` file (default is `5000`). You can access the API at `http://localhost:5000`.

## API Documentation

The API documentation is available via Swagger. You can access it at: `http://localhost:5000/swagger/index.html`

## Project Structure

. ├── docs # Swagger documentation files ├── handlers # Handler functions for API endpoints ├── models # Data models ├── routes # API route definitions ├── .env # Environment variables ├── go.mod # Go modules file ├── go.sum # Go modules dependencies file ├── main.go # Main application file └── README.md # This file

## Environment Variables

The application uses the following environment variables:

- `MONGODB_URI`: The URI for connecting to MongoDB.
- `PORT`: The port on which the server will run.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any changes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
