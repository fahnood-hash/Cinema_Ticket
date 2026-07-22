# Cinema Ticket Booking System

A real-time cinema seat-booking system designed to prevent double booking when multiple users select the same seat simultaneously.

Built with Go, Vue 3, MongoDB, Redis, RabbitMQ, Firebase Authentication, WebSocket, and Docker Compose.

## Table of Contents

1. [System Architecture](#system-architecture)
2. [Tech Stack](#tech-stack)
3. [Booking Flow](#booking-flow)
4. [Redis Lock Strategy](#redis-lock-strategy)
5. [Message Queue](#message-queue-rabbitmq)
6. [Audit Logs](#audit-logs)
7. [How to Run](#how-to-run)
8. [Assumptions and Trade-offs](#assumptions-and-trade-offs)
9. [Future Improvements](#future-improvements)

---

## System Architecture

<img width="1232" height="1000" alt="System architecture diagram" src="https://github.com/user-attachments/assets/081a40a5-519a-4f2f-b2d3-4694ae660bd1" />

The Vue client communicates with the Go API through HTTP and WebSocket. Redis handles temporary distributed seat locks, MongoDB stores confirmed bookings, and RabbitMQ sends booking events to a background worker for audit logging.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go, Gin |
| Frontend | Vue 3, Vite |
| Authentication | Firebase Authentication with Google Sign-In |
| Database | MongoDB |
| Distributed Lock | Redis |
| Message Queue | RabbitMQ |
| Realtime | WebSocket |
| Background Worker | Go |
| Containerization | Docker Compose |

---

## Booking Flow

### User Booking Flow

1. User signs in with Google.
2. Firebase returns a unique Firebase User ID.
3. User opens the seat map.
4. User selects an available seat.
5. Backend locks the seat in Redis for five minutes.
6. Other users cannot reserve the same seat.
7. User confirms the booking. Payment is mocked.
8. Booking is saved into MongoDB.
9. RabbitMQ publishes a `BOOKING_SUCCESS` event.
10. The background worker stores an audit log.
11. The seat becomes `BOOKED`.

### Timeout Flow

1. Redis lock expires after five minutes.
2. The seat becomes `AVAILABLE`.
3. WebSocket notifies connected clients.
4. RabbitMQ publishes a `BOOKING_TIMEOUT` event.
5. The background worker stores a timeout audit log.

---

## Redis Lock Strategy

Every seat has one Redis key.

Example:

```text
seat:lock:A1
```

The backend uses this Redis command:

```text
SET seat:lock:A1 session-id NX EX 300
```

| Option | Purpose |
|---|---|
| `NX` | Creates the lock only if it does not already exist |
| `EX 300` | Expires the lock after 300 seconds, or five minutes |
| `session-id` | Identifies the user’s temporary booking session |

This guarantees that only one user can reserve a specific seat at a time and prevents double booking.

---

## Message Queue: RabbitMQ

RabbitMQ is used to process booking events asynchronously.

Instead of writing audit logs during the booking request, the backend publishes an event to RabbitMQ:

```text
BOOKING_SUCCESS
BOOKING_TIMEOUT
```

The background worker consumes these events and stores audit logs in MongoDB.

### Benefits

- Faster API response
- Loose coupling between booking and audit-log services
- Easy to extend with email, notifications, or analytics
- Better scalability

---

# Audit Logs

Audit logs are stored asynchronously by the Go background worker.

### Recorded Events

```text
BOOKING_SUCCESS
BOOKING_TIMEOUT
```

### View Audit Logs

```bash
docker compose exec mongo mongosh cinema_booking --eval "db.audit_logs.find().sort({created_at:-1}).pretty()"
```

Each audit record includes:

- Event Type
- Booking ID
- User ID
- Seat ID
- Timestamp
---
## How to Run

### Prerequisites

- Docker Desktop
- Node.js
- Firebase project with Google Sign-In enabled
- Firebase Admin SDK service-account JSON file

### Environment Variables

Create a root `.env` file:

```env
ADMIN_EMAIL=your-email@gmail.com
```

The email configured here receives the `ADMIN` role.

### Firebase Credentials

Place the Firebase Admin SDK credentials here:

```text
backend/secrets/firebase-service-account.json
```

### Start Backend Services

From the project root:

```bash
docker compose up --build
```

This starts the Go API, MongoDB, Redis, RabbitMQ, and the audit-log worker.

### Start Frontend

In another terminal:

```bash
cd frontend
npm install
npm run dev
```

Open the application:

```text
http://localhost:5173
```

---

## Assumptions and Trade-offs

- The system uses one fixed movie.
- The seat layout is fixed from `A1` to `E8`.
- Payment is mocked by the Confirm button.
- MongoDB stores confirmed bookings and audit logs.
- Redis stores temporary seat locks and booking sessions.
- The frontend polls seat data as a fallback if WebSocket is unavailable.

---

## Future Improvements

- Dockerize the frontend for complete one-command startup.
- Protect every booking endpoint with Firebase middleware.
- Add a user booking-history page.
- Add automated concurrency testing.
- Integrate a real payment provider.
- Add email or LINE notifications after booking confirmation.
