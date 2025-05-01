
# Auth Service API

Auth Service API is a RESTful API for user authentication and management. It provides features such as user registration, login, email verification, password reset, and user management.

---

## Features

- **User Registration**: Create a new user account.
- **Login**: Authenticate users and return a JWT token.
- **Email Verification**: Verify user email addresses.
- **Password Reset**: Request and reset passwords securely.
- **User Management**: Retrieve and update user details.
- **Admin Features**: Access all users (admin-only).

---

## Technologies Used

- **Programming Language**: Go (Golang)
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL
- **ORM**: pgx (PostgreSQL driver)
- **Authentication**: JWT (JSON Web Tokens)
- **Documentation**: Swagger (via Swaggo)

---

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/cevrimxe/auth-service.git
   cd auth-service
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up the environment variables:
   Create a `.env` file in the root directory and add the following:
   ```env
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=your_db_user
   DB_PASSWORD=your_db_password
   DB_NAME=your_db_name
   SMTP_HOST=smtp.example.com
   SMTP_PORT=587
   SMTP_SENDER_EMAIL=your_email@example.com
   SMTP_SENDER_PASSWORD=your_email_password
   ```

4. Run database migrations (if applicable).

5. Start the server:
   ```bash
   go run cmd/main.go
   ```

6. Access the API at:
   ```
   http://localhost:8080
   ```

7. Access Swagger documentation at:
   ```
   http://localhost:8080/docs
   ```

---

## API Endpoints

### Auth Endpoints

| Method | Endpoint          | Description                          |
|--------|-------------------|--------------------------------------|
| POST   | `/signup`         | Register a new user                 |
| POST   | `/login`          | Log in a user and return a JWT token|
| GET    | `/verify`         | Verify user email using a token     |
| POST   | `/forgot-password`| Request a password reset            |
| POST   | `/reset-password` | Reset a user's password             |

### User Endpoints

| Method | Endpoint          | Description                          |
|--------|-------------------|--------------------------------------|
| GET    | `/me`             | Get the authenticated user's details|
| PUT    | `/me`             | Update the authenticated user's details|
| PUT    | `/change-password`| Change the authenticated user's password|

### Admin Endpoints

| Method | Endpoint          | Description                          |
|--------|-------------------|--------------------------------------|
| GET    | `/users`          | Retrieve a list of all users (admin-only)|

---

## Environment Variables

| Variable             | Description                          |
|----------------------|--------------------------------------|
| `DB_HOST`            | Database host (e.g., localhost)     |
| `DB_PORT`            | Database port (e.g., 5432)          |
| `DB_USER`            | Database username                   |
| `DB_PASSWORD`        | Database password                   |
| `DB_NAME`            | Database name                       |
| `SMTP_HOST`          | SMTP server host (e.g., smtp.gmail.com) |
| `SMTP_PORT`          | SMTP server port (e.g., 587)        |
| `SMTP_SENDER_EMAIL`  | Email address used for sending emails |
| `SMTP_SENDER_PASSWORD` | Password for the sender email account |


---

## Project Structure

```
auth-service/
│
├── cmd/
│   └── main.go          # Entry point of the application
│
├── config/
│   └── config.go        # Configuration management
│
├── database/
│   └── database.go      # Database connection setup
│
├── docs/
│   └── docs.go          # Swagger documentation (generated)
│
├── models/
│   ├── user.go          # User model and database operations
│   └── email.go         # Email-related logic (e.g., sending emails)
│
├── routes/
│   ├── routes.go        # Route registration
│   └── users.go         # User-related endpoints
│
├── middlewares/
│   └── auth.go          # Authentication middleware
│
├── utils/
│   ├── utils.go         # Utility functions (e.g., JWT, general helpers)
│   └── hash.go          # Password hashing and verification
│
├── .env                 # Environment variables
├── go.mod               # Go module file
├── go.sum               # Go dependencies
└── README.md            # Project documentation
```

---

## How to Use Swagger

1. Install Swaggo CLI:
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

2. Generate Swagger documentation:
   ```bash
   swag init --dir ./cmd,./routes,./models,./config,./database --output ./docs
   ```

3. Access Swagger UI at:
   ```
   http://localhost:8080/docs
   ```

---

## Testing

1. Run unit tests:
   ```bash
   go test ./...
   ```

2. Use tools like Postman or Swagger UI to test API endpoints.

---

## Future Improvements

- Add Two-Factor Authentication (2FA).
- Implement rate limiting for sensitive endpoints.
- Add email templates for better user experience.
- Improve error handling and logging.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Contact

For support or inquiries, please contact:
- **Name**: Ahmet
- **Email**: cevrimdev@gmail.com



