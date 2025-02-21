# Bookaroo Platform Backend

A Go-based backend service for the Bookaroo Platform, providing property listing and booking functionalities.

## Features

- Property listing and details
- Property search functionality
- Booking management
- User dashboard for property owners and guests

## Prerequisites

- Go 1.21 or higher
- PostgreSQL
- Git

## Setup

1. Clone the repository:
```bash
git clone https://github.com/bookaroo/bookaroo-platform-be.git
cd bookaroo-platform-be
```

2. Install dependencies:
```bash
go mod download
```

## Environment Setup

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Update the `.env` file with your configuration:
   - Database credentials
   - Server configuration
   - JWT settings
   - AWS credentials (if using S3 for image storage)
   - Email settings (if implementing email notifications)
   - Redis configuration (if implementing caching)

3. Important Security Notes:
   - Never commit the `.env` file to version control
   - Keep your JWT secret secure and unique for each environment
   - Regularly rotate API keys and access credentials
   - Use strong passwords for database and service accounts

## Database Setup

1. Create the database:
```bash
createdb bookaroo
```

## Run the Application

1. Run the application:
```bash
go run main.go
```

The server will start at `http://localhost:8080`

## API Endpoints

### Properties
- `GET /api/properties` - List all properties
- `GET /api/properties/:id` - Get property details
- `GET /api/properties/search` - Search properties with filters
- `POST /api/properties` - Create a new property
- `PUT /api/properties/:id` - Update an existing property
- `GET /api/properties/:id/owner-details` - Get detailed property information for owners (includes booking status and history)

### Bookings
- `POST /api/bookings` - Create a new booking
- `GET /api/bookings/guest/:guest_id` - Get list of bookings for a guest user (includes booking history and statistics)

### User Dashboard
- `GET /api/dashboard` - Get user dashboard (different view for owners and guests)

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request
