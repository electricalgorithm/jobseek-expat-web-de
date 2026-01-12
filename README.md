# Expatter Backend

Backend API server for Expatter Jobs - A job search aggregator helping English-speaking expats find opportunities in Europe.

## Overview

Expatter is a specialized job search platform that:
- Aggregates job postings from LinkedIn and Indeed
- Filters for English-speaking positions in European countries
- Removes jobs requiring local language fluency
- Provides email alerts for new matching jobs (Pro feature)
- Tracks job history to prevent duplicate notifications

## Tech Stack

- **Language**: Go 1.23+
- **Database**: SQLite3
- **Authentication**: JWT (HS256)
- **Job Scraper**: `jobseek-expat` (Python CLI)
- **Email Service**: Resend
- **Scheduler**: robfig/cron
- **Frontend Serving**: Static file server

## Features

### Core Functionality
- **Job Search**: Real-time job search across multiple platforms
- **Smart Filtering**: 
  - Language detection (English-only)
  - Local language requirement filtering
  - Keyword exclusion
  - Site selection (LinkedIn, Indeed)
  - Time-based filtering (hours old)
- **User Management**:
  - Registration with trial period (7 days)
  - JWT-based authentication
  - Subscription plans (Basic/Pro)
- **Email Alerts** (Pro):
  - Scheduled job searches
  - Customizable frequency (hourly/daily)
  - Duplicate prevention
  - Unsubscribe functionality
- **Payment Integration**: Mock payment verification system

## Project Structure

```
jobseek-web-be/
├── main.go                 # Entry point
├── internal/
│   ├── auth/              # Authentication & JWT
│   ├── db/                # Database setup & migrations
│   ├── email/             # Email templates & sending
│   ├── handlers/          # HTTP handlers
│   ├── models/            # Data models
│   ├── scheduler/         # Cron job scheduler
│   └── search/            # Job search service
├── data/                  # SQLite database (gitignored)
├── Dockerfile             # Production Docker image
└── docker-compose.yml     # Docker setup
```

## Setup

### Prerequisites

- **Go**: 1.23 or higher
- **Python**: 3.12+ (for jobseek-expat CLI)
- **Node.js**: 18+ (for frontend, optional)

### Local Development

1. **Clone the repository**:
   ```bash
   git clone <repo-url>
   cd jobseek-web-be
   ```

2. **Install jobseek-expat CLI**:
   ```bash
   python3 -m venv .venv
   source .venv/bin/activate  # On Windows: .venv\Scripts\activate
   pip install jobseek-expat
   ```

3. **Install Go dependencies**:
   ```bash
   go mod download
   ```

4. **Set up environment variables**:
   Create a `.env` file (optional, can use exports):
   ```env
   PORT=8080
   APP_NAME=Expatter
   APP_DOMAIN=http://localhost:8080
   RESEND_API_KEY=your_resend_api_key
   EMAIL_FROM=noreply@yourdomain.com
   SCHEDULER_FREQUENCY=@every 1h
   DB_PATH=./data/jobseek.db
   FRONTEND_PATH=../jobseek-web-fe/dist
   ```

5. **Run the server**:
   ```bash
   go run main.go
   ```

   The server will start on `http://localhost:8080`.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `APP_NAME` | Application name | `Expatter` |
| `APP_DOMAIN` | Full domain URL | `http://localhost:8080` |
| `RESEND_API_KEY` | Resend API key for emails | (required for email) |
| `EMAIL_FROM` | Sender email address | `onboarding@resend.dev` |
| `SCHEDULER_FREQUENCY` | Cron schedule for alerts | `@every 1h` |
| `DB_PATH` | SQLite database path | `./data/jobseek.db` |
| `FRONTEND_PATH` | Frontend static files path | `../jobseek-web-fe/dist` |

## Database Schema

### `users`
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,  -- bcrypt hashed
    subscription_plan TEXT DEFAULT 'basic',
    paid INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### `user_searches`
```sql
CREATE TABLE user_searches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    keyword TEXT,
    country TEXT,
    location TEXT,
    language TEXT,
    frequency TEXT DEFAULT 'hourly',
    hours_old INTEGER DEFAULT 24,
    exclude TEXT DEFAULT '',
    results_wanted INTEGER DEFAULT 10,
    last_run DATETIME,
    FOREIGN KEY(user_id) REFERENCES users(id)
);
```

### `sent_jobs`
```sql
CREATE TABLE sent_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    search_id INTEGER NOT NULL,
    job_url TEXT NOT NULL,
    sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(search_id) REFERENCES user_searches(id),
    UNIQUE(search_id, job_url)
);
```

## API Endpoints

### Authentication

#### POST `/api/register`
Register a new user (starts trial period).

**Request**:
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepass123",
  "subscription": "basic"
}
```

**Response**: `201 Created`
```json
{
  "message": "User registered successfully"
}
```

#### POST `/api/login`
Authenticate and receive JWT token.

**Request**:
```json
{
  "email": "john@example.com",
  "password": "securepass123"
}
```

**Response**: `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "name": "John Doe",
  "email": "john@example.com",
  "subscription": "basic"
}
```

### Job Search

#### POST `/api/search`
Search for jobs (requires authentication).

**Headers**:
```
Authorization: Bearer <token>
```

**Request**:
```json
{
  "keyword": "Software Engineer",
  "country": "Germany",
  "location": "Berlin, Remote",
  "local_language": "German",
  "hours_old": 24,
  "exclude": "senior, lead",
  "results_wanted": 20
}
```

**Response**: `200 OK`
```json
[
  {
    "site": "linkedin",
    "title": "Software Engineer",
    "company": "Tech Company",
    "location": "Berlin, Germany",
    "date_posted": "2026-01-12",
    "job_url": "https://...",
    "job_level": "mid_senior"
  }
]
```

### Saved Searches (Pro Only)

#### POST `/api/searches`
Save a search for email alerts.

**Request**:
```json
{
  "keyword": "Software Engineer",
  "country": "Germany",
  "location": "Remote",
  "language": "German",
  "frequency": "hourly",
  "hours_old": 24,
  "exclude": "senior",
  "results_wanted": 30
}
```

**Response**: `201 Created`

#### GET `/api/searches`
List all saved searches for authenticated user.

**Response**: `200 OK`
```json
[
  {
    "id": 1,
    "keyword": "Software Engineer",
    "country": "Germany",
    "location": "Remote",
    "frequency": "hourly",
    "last_run": "2026-01-12T10:00:00Z"
  }
]
```

#### DELETE `/api/searches/:id`
Delete a saved search.

**Response**: `200 OK`

### Payment

#### POST `/api/payment/verify`
Verify payment and upgrade user.

**Request**:
```json
{
  "email": "john@example.com",
  "token": "tok_stripe_xxx",
  "subscription": "pro"
}
```

**Response**: `200 OK`

### Utility

#### GET `/api/redirect?data=<base64_url>`
Track job clicks and redirect to actual job URL.

#### GET `/unsubscribe?uid=<user_id>&sid=<search_id>`
Unsubscribe from email alerts.

## Docker Deployment

### Build and Run

```bash
# Build image
docker compose build

# Run in background
docker compose up -d

# View logs
docker compose logs -f

# Stop
docker compose down
```

### Production Deployment

The application is designed to run on ARM64 (Raspberry Pi) or AMD64 architectures.

**Environment Setup**:
```bash
# On deployment server
export RESEND_API_KEY="re_xxx"
export APP_DOMAIN="https://yourapp.com"
export EMAIL_FROM="jobs@yourapp.com"

# Pull and run
git pull origin main
docker compose build
docker compose up -d
```

## Development

### Running Tests
```bash
go test ./...
```

### Code Structure

- **Handlers**: HTTP route handlers in `internal/handlers/`
- **Services**: Business logic in respective packages
- **Models**: Data structures in `internal/models/`
- **Database**: Schema and initialization in `internal/db/`

### Adding New Features

1. Add models in `internal/models/`
2. Update database schema in `internal/db/db.go`
3. Create handler in `internal/handlers/`
4. Register route in `main.go`

## Scheduler

The application runs a background scheduler that:
- Checks saved searches based on frequency
- Executes job searches via `jobseek-expat` CLI
- Filters out previously sent jobs
- Sends email notifications
- Updates search history

**Frequency Options**:
- `hourly`: Runs every hour
- `daily`: Runs once per day

## Security

- **Passwords**: Hashed using bcrypt (cost 10)
- **JWT**: HS256 algorithm with secret key
- **Trial Period**: 7 days from registration
- **Rate Limiting**: Consider implementing for production

## Troubleshooting

### jobseek-expat not found
Ensure Python virtual environment is activated or `jobseek-expat` is in PATH:
```bash
source .venv/bin/activate
which jobseek-expat
```

### Database locked
SQLite may lock during concurrent writes. Ensure only one instance is running.

### Email template not found
If running in Docker, ensure the template is embedded via `//go:embed` directive in `internal/email/service.go`.

## License

[Your License Here]

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## Support

For issues and questions, please open an issue on GitHub.
