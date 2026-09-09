# Ticket System (Golang Backend Intern Assignment)

A small backend service where users can register, log in, create tickets,
view their own tickets, and update the status of their own tickets.

> Status: work in progress — this README will be filled in as the project
> is built (local run instructions, Docker instructions, deployed URL,
> and assumptions made).

## Tech Stack
- Language: Go 1.22
- Storage: SQLite (to be added)
- Auth: JWT (to be added)

## Endpoints (per assignment contract)
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | /health | Health check |
| POST | /auth/register | Register user |
| POST | /auth/login | Login and return JWT |
| POST | /tickets | Create ticket |
| GET | /tickets | List logged-in user's tickets |
| GET | /tickets/{id} | Get own ticket by ID |
| PATCH | /tickets/{id}/status | Update own ticket status |

## Local Run (once complete)
```
go run main.go
curl http://localhost:8080/health
```

## Docker Run (once complete)
```
docker build -t ticket-system .
docker run -p 8080:8080 ticket-system
curl http://localhost:8080/health
```

## Deployed URL
TBD

## Assumptions
- TBD (will document ownership error codes, status transition rules, etc.)
