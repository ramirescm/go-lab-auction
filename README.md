# Auction System - Automatic Closure Feature ⚡

This project is an extension of the FullCycle Auction system. The primary objective was to implement a background routine that automatically closes an auction after a specific duration defined via environment variables.

## 🚀 Features Added

- **Dynamic Duration:** Auction time is calculated based on the `AUCTION_DURATION` environment variable.
- **Non-blocking Closure:** A dedicated Goroutine is spawned upon auction creation to monitor expiration.
- **Concurrency Safe:** The closure process uses context-aware MongoDB updates to ensure data integrity.
- **Automated Testing:** Included logic to verify that auctions transition from `Active` to `Completed` status without manual intervention.

---

## 🛠 Tech Stack

- **Language:** Go (Golang)
- **Database:** MongoDB
- **Concurrency:** Goroutines & Channels
- **Containerization:** Docker & Docker Compose
- **Tracing:** OpenTelemetry (Optional integration)

---

## 📋 Environment Variables

Before running the application, ensure your `.env` file (located in `cmd/auction/.env`) contains the following key:

| Variable           | Description                                    | Example                               |
| ------------------ | ---------------------------------------------- | ------------------------------------- |
| `AUCTION_DURATION` | The time until an auction closes automatically | `30s`, `1m`, `2h`                     |
| `MONGODB_URL`      | Connection string for MongoDB                  | `mongodb://admin:admin@mongodb:27017` |
| `MONGODB_DB`       | Database name                                  | `auctions`                            |

---

## 🏃 How to Run in Development

The easiest way to get the environment up and running is using Docker Compose.

### 1. Build and Start the Containers

From the root directory, run:

```bash
docker-compose up -d --build

```

### 2. Verify Connectivity

Ensure MongoDB is running on port `27017` and the App is on `8080`.

- **MongoDB:** `localhost:27017`
- **API Entry Point:** `localhost:8080`

---

## 🧪 Running Tests

To validate the automatic closure, a test case was developed to simulate the passage of time and check the database status.

### Run tests locally:

Make sure you have a local MongoDB instance or use the one in Docker (adjust `localhost` in your connection string if needed):

```bash
go test ./internal/infra/database/auction/... -v

```

---

## 🏗 Implementation Details

The core logic resides in `internal/infra/database/auction/create_auction.go`.

1. **Creation:** When `CreateAuction` is called, the data is persisted in MongoDB.
2. **Goroutine Spawn:** Immediately after a successful insertion, a new Goroutine is started: `go ar.waitAndCloseAuction(ctx, auctionId)`.
3. **Sleeping:** The routine calculates the duration via `os.Getenv` and calls `time.Sleep`.
4. **Update:** Once the timer expires, an `UpdateOne` command sets the auction status to `Completed` (1).

---

## 📨 API Testing (REST Client)

You can use the following `.http` template in VS Code to test the feature:

```http
### Create a New Auction
POST http://localhost:8080/auction
Content-Type: application/json

{
    "product_name": "Vintage Camera",
    "category": "Photography",
    "description": "A well-preserved 1950s camera.",
    "condition": 1
}

### Wait for the duration defined in AUCTION_DURATION...

### Check if it closed automatically
GET http://localhost:8080/auction/{{auction_id}}

```
