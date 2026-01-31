# Backend Setup Walkthrough

I have successfully initialized the **IOI AMMS** backend using **Go** and **PostgreSQL**.

## Architecture
- **Language:** Go 1.25+ (Alpine)
- **Database:** PostgreSQL 16
- **Routing:** Chi Router
- **Project Name:** `ioi_amms` (Explicitly set)
- **Containerization:** Docker Compose
    - Service: `api` -> Container: `IOI_AMMS`
    - Service: `db` -> Container: `postgres_db`
    - Service: `minio` -> Container: `minio`

## Credentials

### PostgreSQL
- **Host:** `localhost`
- **Port:** `5432`
- **User:** `postgres`
- **Password:** `postgres`
- **Database:** `ioi_amms`

### MinIO
- **Console:** [http://localhost:9001](http://localhost:9001)
- **User:** `minioadmin`
- **Password:** `minioadmin`

## Verification Results

### 1. Containers & Network
The project runs as `ioi_amms`.
- **Containers:** `IOI_AMMS`, `postgres_db`, `minio`
- **Volume:** `ioi_amms_postgres_data`, `ioi_amms_minio_data`

### 2. Database Schema
I have applied the schema derived from the PRD, creating tables for:
- Tenants & Users
- Locations & Assets
- Work Orders & Maintenance
- Inventory & Logistics

### 3. API Health Check
The backend is reachable at `http://localhost:8080`.

**Request:**
```bash
curl -v http://localhost:8080/health
```

**Response:**
```json
{
  "message": "It's healthy",
  "status": "up"
}
```

## How to Run
If you need to restart the services:
```bash
docker-compose up -d --build
```
The backend supports hot-reloading via Air, so you can edit code in `backend/` and it will auto-rebuild.
