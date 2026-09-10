# ---------- Build stage ----------
# Using a Go version that matches (or exceeds) what's in go.mod.
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copy dependency files first so Docker can cache this layer —
# it only re-downloads modules when go.mod/go.sum actually change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 gives a fully static binary (no external C libraries needed).
# This works here because modernc.org/sqlite is a pure-Go SQLite driver.
RUN CGO_ENABLED=0 GOOS=linux go build -o ticket-system .

# ---------- Final stage ----------
# A minimal image containing only the compiled binary — much smaller
# and more secure than shipping the whole Go toolchain.
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/ticket-system .

EXPOSE 8080

CMD ["./ticket-system"]
