# IOI AMMS API Specification v1

> **Last Updated:** 2026-01-31  
> **Base URL:** `/api/v1`  
> **Content-Type:** `application/json`

---

## 1. Authentication

All protected endpoints require a Bearer token in the `Authorization` header.

### Headers
```
Authorization: Bearer <jwt_token>
```

### `POST /auth/login`
Exchange credentials for access and refresh tokens.

**Request:**
```json
{
  "email": "admin@ioi.com",
  "password": "password123"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "00000000-0000-0000-0000-000000000002",
    "email": "admin@ioi.com",
    "name": "Admin User",
    "role": "manager",
    "tenantId": "00000000-0000-0000-0000-000000000001"
  }
}
```

**Error (401 Unauthorized):**
```json
{
  "error": "Unauthorized",
  "message": "Invalid credentials"
}
```

### `POST /auth/refresh`
Refresh an expired access token.

**Request:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

## 2. Roles & Permissions (RBAC)

### Role Hierarchy
| Role | Description | Access Level |
|------|-------------|--------------|
| `technician` | Field technician | Execute assigned WOs, view assets |
| `supervisor` | Team supervisor | Assign WOs, verify assets, view reports |
| `storeman` | Inventory manager | Manage inventory and transfers |
| `manager` | Operations manager | Full access except admin functions |
| `admin` | System administrator | Full system access |

### Role → Permission Matrix
| Permission | technician | supervisor | storeman | manager | admin |
|------------|:----------:|:----------:|:--------:|:-------:|:-----:|
| `asset:read` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `asset:write` | | ✓ | | ✓ | ✓ |
| `asset:delete` | | | | ✓ | ✓ |
| `wo:read` | ✓ | ✓ | | ✓ | ✓ |
| `wo:write` | ✓ | ✓ | | ✓ | ✓ |
| `wo:assign` | | ✓ | | ✓ | ✓ |
| `wo:close` | | ✓ | | ✓ | ✓ |
| `inventory:read` | | | ✓ | ✓ | ✓ |
| `inventory:write` | | | ✓ | ✓ | ✓ |
| `report:view` | | ✓ | | ✓ | ✓ |
| `user:manage` | | | | ✓ | ✓ |

---

## 3. Common Patterns

### Pagination
All list endpoints support pagination.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number (1-indexed) |
| `limit` | int | 10 | Items per page (max 100) |
| `sortBy` | string | `created_at` | Field to sort by |
| `sortDir` | string | `desc` | Sort direction: `asc` or `desc` |

**Response Wrapper:**
```json
{
  "data": [...],
  "meta": {
    "total": 150,
    "page": 1,
    "limit": 10,
    "totalPages": 15
  }
}
```

### Error Responses
All errors follow this format:
```json
{
  "error": "HTTP Status Text",
  "code": "MACHINE_READABLE_CODE",
  "message": "Human-readable description",
  "details": { "field": "additional context" }
}
```

**Error Codes:**
| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Invalid input data |
| `NOT_FOUND` | Resource doesn't exist |
| `UNAUTHORIZED` | Authentication required |
| `FORBIDDEN` | Insufficient permissions |
| `CONFLICT` | Resource state conflict |
| `RATE_LIMITED` | Too many requests |
| `FILE_TOO_LARGE` | Upload exceeds 10MB limit |
| `DATABASE_ERROR` | Database operation failed |
| `INTERNAL_ERROR` | Unexpected server error |

**HTTP Status Codes:**
| Status | Error | Common Causes |
|--------|-------|---------------|
| 400 | Bad Request | Invalid JSON, missing required fields |
| 401 | Unauthorized | Missing/invalid/expired token |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 413 | Payload Too Large | File exceeds 10MB |
| 429 | Too Many Requests | Rate limit exceeded (100 req/min) |
| 500 | Internal Server Error | Server-side error |

### Rate Limiting
- **Limit:** 100 requests per minute per IP
- **Headers:** `Retry-After: 60` on 429 responses

---

## 4. Assets API

### Asset Object Schema
```typescript
interface Asset {
  id: string;           // UUID
  tenantId: string;     // UUID
  parentId?: string;    // UUID (for asset hierarchy)
  locationId?: string;  // UUID
  name: string;
  status: AssetStatus;
  criticality: AssetCriticality;
  lastInspection?: string;  // ISO 8601 datetime
  createdAt: string;    // ISO 8601 datetime
  updatedAt: string;    // ISO 8601 datetime
  location?: string;    // Computed: location name
}

type AssetStatus = 'operational' | 'maintenance' | 'decommissioned' | 'pending';
type AssetCriticality = 'high' | 'medium' | 'low';
```

### `GET /assets`
List all assets with filtering and pagination.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `status` | string[] | Filter by status (repeatable) |
| `criticality` | string[] | Filter by criticality (repeatable) |
| `search` | string | Search in name |
| `page` | int | Page number |
| `limit` | int | Items per page |
| `sortBy` | string | Sort field |
| `sortDir` | string | `asc` or `desc` |

**Example Request:**
```
GET /api/v1/assets?status=operational&status=maintenance&criticality=high&page=1&limit=20
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
      "tenantId": "00000000-0000-0000-0000-000000000001",
      "parentId": null,
      "locationId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "name": "Generator X-500",
      "status": "operational",
      "criticality": "high",
      "lastInspection": "2026-01-15T10:30:00Z",
      "createdAt": "2025-06-01T08:00:00Z",
      "updatedAt": "2026-01-15T10:30:00Z",
      "location": "Building A"
    }
  ],
  "meta": {
    "total": 1,
    "page": 1,
    "limit": 20,
    "totalPages": 1
  }
}
```

### `GET /assets/{id}`
Get single asset by ID.

**Response (200 OK):** Single Asset object

**Response (404 Not Found):**
```json
{
  "error": "Not Found",
  "message": "Asset not found"
}
```

### `POST /assets`
Create a new asset.

**Request Body:**
```json
{
  "name": "New Generator",
  "parentId": null,
  "locationId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "pending",
  "criticality": "medium"
}
```

| Field | Type | Required | Default |
|-------|------|----------|---------|
| `name` | string | ✓ | - |
| `parentId` | string | | null |
| `locationId` | string | | null |
| `status` | string | | `pending` |
| `criticality` | string | | `medium` |

**Response (201 Created):** Created Asset object

### `PUT /assets/{id}`
Update an existing asset.

**Request Body:** Same as POST

**Response (200 OK):** Updated Asset object

### `DELETE /assets/{id}`
Delete an asset.

**Response (204 No Content):** Success, no body

---

## 5. Locations API

### Location Object Schema
```typescript
interface Location {
  id: string;           // UUID
  tenantId: string;     // UUID
  parentId?: string;    // UUID (for hierarchy)
  name: string;
  type: LocationType;
  createdAt: string;
  updatedAt: string;
}

type LocationType = 'Site' | 'Building' | 'Room' | 'Zone';
```

### `GET /locations`
List all locations for the tenant.

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "tenantId": "00000000-0000-0000-0000-000000000001",
      "parentId": null,
      "name": "Main Campus",
      "type": "Site",
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ]
}
```

### `GET /locations/{id}`
Get single location.

### `GET /locations/{id}/children`
Get child locations of a parent.

**Response (200 OK):**
```json
{
  "data": [
    { "id": "...", "name": "Building A", "type": "Building", ... },
    { "id": "...", "name": "Building B", "type": "Building", ... }
  ]
}
```

### `POST /locations`
Create a new location.

**Request Body:**
```json
{
  "name": "Building A",
  "parentId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "type": "Building"
}
```

| Field | Type | Required |
|-------|------|----------|
| `name` | string | ✓ |
| `type` | string | ✓ |
| `parentId` | string | |

### `PUT /locations/{id}`
Update a location.

### `DELETE /locations/{id}`
Delete a location.

---

## 6. Work Orders API

### Work Order Object Schema
```typescript
interface WorkOrder {
  id: string;           // UUID
  tenantId: string;     // UUID
  assetId?: string;     // UUID
  status: WOStatus;
  origin: WOOrigin;
  priority: WOPriority;
  description?: string;
  createdAt: string;
  updatedAt: string;
  assetName?: string;   // Computed: asset name
}

type WOStatus = 'Draft' | 'Ready' | 'In_Progress' | 'Closed';
type WOOrigin = 'PM' | 'CM' | 'Defect';
type WOPriority = 'Low' | 'Medium' | 'High' | 'Critical';
```

### `GET /work-orders`
List work orders with filtering.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `status` | string[] | Filter by status (repeatable) |
| `priority` | string[] | Filter by priority (repeatable) |
| `assetId` | string | Filter by asset |
| `page` | int | Page number |
| `limit` | int | Items per page |

**Example:**
```
GET /api/v1/work-orders?status=Draft&status=Ready&priority=High
```

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "wo-uuid",
      "tenantId": "tenant-uuid",
      "assetId": "asset-uuid",
      "status": "Draft",
      "origin": "CM",
      "priority": "High",
      "description": "Noise from engine compartment",
      "createdAt": "2026-01-31T10:00:00Z",
      "updatedAt": "2026-01-31T10:00:00Z",
      "assetName": "Generator X-500"
    }
  ],
  "meta": { "total": 1, "page": 1, "limit": 10, "totalPages": 1 }
}
```

### `GET /work-orders/{id}`
Get single work order.

### `POST /work-orders`
Create a new work order.

**Request Body:**
```json
{
  "assetId": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "origin": "CM",
  "priority": "High",
  "description": "Engine making unusual noise"
}
```

| Field | Type | Required | Default |
|-------|------|----------|---------|
| `origin` | string | ✓ | - |
| `assetId` | string | | null |
| `priority` | string | | `Medium` |
| `description` | string | | null |

**Response (201 Created):** Created work order with `status: "Draft"`

### `PUT /work-orders/{id}`
Update a work order.

**Request Body:**
```json
{
  "priority": "Critical",
  "description": "Updated description"
}
```

### `PATCH /work-orders/{id}/status`
Update work order status (state transition).

**Request Body:**
```json
{
  "status": "In_Progress"
}
```

**Response (200 OK):**
```json
{
  "status": "In_Progress"
}
```

---

## 7. Enums Reference

### Asset Status
| Value | Display | Color |
|-------|---------|-------|
| `operational` | Operational | Green |
| `maintenance` | Under Maintenance | Yellow |
| `decommissioned` | Decommissioned | Red |
| `pending` | Pending | Blue |

### Asset Criticality
| Value | Display | Color |
|-------|---------|-------|
| `high` | High | Red |
| `medium` | Medium | Orange |
| `low` | Low | Gray |

### Work Order Status
| Value | Display | Next States |
|-------|---------|-------------|
| `Draft` | Draft | Ready |
| `Ready` | Ready | In_Progress |
| `In_Progress` | In Progress | Closed |
| `Closed` | Closed | - |

### Work Order Origin
| Value | Display | Description |
|-------|---------|-------------|
| `PM` | Preventive | Scheduled maintenance |
| `CM` | Corrective | Reactive/breakdown |
| `Defect` | Defect | From defect report |

### Work Order Priority
| Value | Display | SLA |
|-------|---------|-----|
| `Low` | Low | 72h |
| `Medium` | Medium | 48h |
| `High` | High | 24h |
| `Critical` | Critical | 4h |

### Location Type
| Value | Hierarchy Level |
|-------|-----------------|
| `Site` | 1 (Top) |
| `Building` | 2 |
| `Room` | 3 |
| `Zone` | 4 |

---

## 8. Files API (MinIO Storage)

File uploads are stored in MinIO object storage, organized by tenant and asset.

### Upload Result Schema
```typescript
interface UploadResult {
  success: boolean;
  objectName: string;      // Full path in storage
  size: number;            // Bytes
  contentType: string;
  originalName: string;
  downloadUrl: string;     // Relative URL for download
}
```

### `POST /upload`
Upload a file (multipart form-data).

**Request:**
- Content-Type: `multipart/form-data`
- Max Size: **10 MB**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | File | ✓ | The file to upload |
| `assetId` | string | | Associate with asset (default: "general") |

**Example:**
```bash
curl -X POST /api/v1/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@photo.jpg" \
  -F "assetId=asset-uuid"
```

**Response (201 Created):**
```json
{
  "success": true,
  "objectName": "tenants/tenant-id/assets/asset-id/1234567890.jpg",
  "size": 102400,
  "contentType": "image/jpeg",
  "originalName": "photo.jpg",
  "downloadUrl": "/api/v1/files/attachments/tenants/tenant-id/assets/asset-id/1234567890.jpg"
}
```

**Error (413 Payload Too Large):**
```json
{
  "error": "Request Entity Too Large",
  "code": "FILE_TOO_LARGE",
  "message": "File exceeds maximum size of 10 MB",
  "details": { "maxSizeBytes": 10485760 }
}
```

### `GET /files/{bucket}/{objectPath}`
Download a file (redirects to presigned URL).

**Response:** 307 Temporary Redirect to MinIO presigned URL (valid 1 hour)

### `DELETE /files/{bucket}/{objectPath}`
Delete a file. Only files owned by your tenant can be deleted.

**Response (204 No Content):** Success

**Error (403 Forbidden):**
```json
{
  "error": "Forbidden",
  "code": "FORBIDDEN",
  "message": "You can only delete files owned by your tenant"
}
```

---

## 9. Health & System

### `GET /health`
Health check endpoint (public, no auth required).

**Response (200 OK):**
```json
{
  "status": "up",
  "message": "It's healthy"
}
```

### `GET /`
Root endpoint (public).

**Response (200 OK):**
```json
{
  "message": "IOI AMMS API"
}
```

---

## 9. Frontend Integration Notes

### TypeScript Types
```typescript
// Use these in your frontend for type safety
type UUID = string;
type ISODateTime = string;

interface PaginatedResponse<T> {
  data: T[];
  meta: {
    total: number;
    page: number;
    limit: number;
    totalPages: number;
  };
}

interface ApiError {
  error: string;
  code?: string;           // Machine-readable error code
  message: string;
  details?: Record<string, unknown>;
}
```

### Fetch Example
```typescript
const fetchAssets = async (token: string, filters?: AssetFilters) => {
  const params = new URLSearchParams();
  if (filters?.status) filters.status.forEach(s => params.append('status', s));
  if (filters?.page) params.set('page', String(filters.page));
  
  const res = await fetch(`/api/v1/assets?${params}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  if (!res.ok) {
    const err: ApiError = await res.json();
    throw new Error(err.message);
  }
  
  return res.json() as Promise<PaginatedResponse<Asset>>;
};
```

### Nuxt Composable Pattern
```typescript
// composables/useApi.ts
export const useAssets = () => {
  const { token } = useAuth();
  
  return useFetch<PaginatedResponse<Asset>>('/api/v1/assets', {
    headers: { Authorization: `Bearer ${token.value}` },
    key: 'assets'
  });
};
```
