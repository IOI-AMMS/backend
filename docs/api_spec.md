# IOI AMMS API Specification v1.1

> **Last Updated:** 2026-02-01  
> **Schema Version:** v1.1  
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
    "role": "Admin",
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

## 2. User Management (Admin/Manager)

### User Object Schema (v1.1)
```typescript
interface User {
  id: string;           // UUID
  tenantId: string;     // UUID
  orgUnitId?: string;   // UUID - Org unit assignment
  email: string;
  fullName?: string;    // Single name field (replaces firstName/lastName)
  role: UserRole;
  isActive: boolean;    // Active status flag
  createdAt: string;
}

type UserRole = 'Technician' | 'Supervisor' | 'Storeman' | 'Manager' | 'Admin' | 'Viewer';
```

### `GET /users`
List all users in the tenant.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `role` | string | Filter by role |
| `page` | int | Page number (default: 1) |
| `limit` | int | Items per page (default: 10) |

**Response (200 OK):**
```json
[
  {
    "id": "uuid...",
    "tenantId": "uuid...",
    "orgUnitId": "uuid...",
    "email": "tech@ioi.com",
    "fullName": "John Technician",
    "role": "Technician",
    "isActive": true,
    "createdAt": "2026-01-01T00:00:00Z"
  }
]
```

### `POST /users`
Create a new user.

**Request Body:**
```json
{
  "email": "newuser@ioi.com",
  "password": "TempPassword123!",
  "role": "Technician",
  "fullName": "Jane Smith",
  "orgUnitId": "uuid-optional"
}
```

**Response (201 Created):** Created User object.

### `PUT /users/{id}/password`
Reset a user's password (admin only).

**Request Body:**
```json
{
  "newPassword": "NewPassword123!"
}
```

**Response (200 OK):** `{"message": "Password updated successfully"}`

---

## 3. Roles & Permissions (RBAC) - v1.1

### Role Hierarchy (Capitalized)
| Role | Description | Access Level |
|------|-------------|--------------|
| `Technician` | Field technician | Execute assigned WOs, view assets |
| `Supervisor` | Team supervisor | Assign WOs, verify assets, view reports |
| `Storeman` | Inventory manager | Manage inventory and transfers |
| `Manager` | Operations manager | Full access except admin functions |
| `Admin` | System administrator | Full system access |
| `Viewer` | Read-only access | View assets and reports only |

### Role → Permission Matrix
| Permission | Technician | Supervisor | Storeman | Manager | Admin |
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
| `tenant:settings` | | | | | ✓ |
| `audit:read` | | | | | ✓ |
| `system:health` | | | | | ✓ |

---

## 4. Common Patterns

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
  "message": "Human-readable description"
}
```

**HTTP Status Codes:**
| Status | Error | Common Causes |
|--------|-------|---------------|
| 400 | Bad Request | Invalid JSON, missing required fields |
| 401 | Unauthorized | Missing/invalid/expired token |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 413 | Payload Too Large | File exceeds 10MB |
| 500 | Internal Server Error | Server-side error |

---

## 5. Assets API (v1.1 Schema)

### Asset Object Schema
```typescript
interface Asset {
  id: string;              // UUID
  tenantId: string;        // UUID
  parentId?: string;       // UUID (for asset hierarchy)
  locationId?: string;     // UUID
  orgUnitId?: string;      // UUID - Organizational unit
  name: string;
  status: AssetStatus;
  isFieldRelated: boolean; // Requires field verification
  isFieldVerified: boolean;// Has been verified in field
  manufacturer?: string;
  modelNumber?: string;
  specs?: object;          // JSONB technical specifications
  createdAt: string;       // ISO 8601 datetime
  updatedAt: string;       // ISO 8601 datetime
  location?: string;       // Computed: location name
  orgUnit?: string;        // Computed: org unit name
}

type AssetStatus = 'Draft' | 'Active' | 'Down' | 'Archived' | 'Red_Tag';
```

### `GET /assets`
List all assets with filtering and pagination.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `status` | string[] | Filter by status (repeatable) |
| `orgUnitId` | string | Filter by org unit |
| `search` | string | Search in name |
| `page` | int | Page number |
| `limit` | int | Items per page |
| `sortBy` | string | Sort field |
| `sortDir` | string | `asc` or `desc` |

**Example Request:**
```
GET /api/v1/assets?status=Active&status=Down&orgUnitId=uuid&page=1&limit=20
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
      "orgUnitId": "uuid...",
      "name": "Generator X-500",
      "status": "Active",
      "isFieldRelated": true,
      "isFieldVerified": true,
      "manufacturer": "Caterpillar",
      "modelNumber": "CAT-500X",
      "specs": {"powerKw": 500, "fuelType": "diesel"},
      "createdAt": "2025-06-01T08:00:00Z",
      "updatedAt": "2026-01-15T10:30:00Z",
      "location": "Building A",
      "orgUnit": "Operations"
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

### `POST /assets`
Create a new asset.

**Request Body:**
```json
{
  "name": "New Generator",
  "parentId": null,
  "locationId": "uuid",
  "orgUnitId": "uuid",
  "status": "Draft",
  "isFieldRelated": true,
  "manufacturer": "Caterpillar",
  "modelNumber": "CAT-300",
  "specs": {"powerKw": 300}
}
```

| Field | Type | Required | Default |
|-------|------|----------|---------|
| `name` | string | ✓ | - |
| `parentId` | string | | null |
| `locationId` | string | | null |
| `orgUnitId` | string | | null |
| `status` | string | | `Draft` |
| `isFieldRelated` | boolean | | true |
| `manufacturer` | string | | null |
| `modelNumber` | string | | null |
| `specs` | object | | {} |

**Response (201 Created):** Created Asset object

### `PUT /assets/{id}`
Update an existing asset.

**Request Body:** Same as POST (all fields optional for partial update)

**Response (200 OK):** Updated Asset object

### `DELETE /assets/{id}`
Delete an asset.

**Response (204 No Content):** Success, no body

---

## 6. Locations API

### Location Object Schema
```typescript
interface Location {
  id: string;           // UUID
  tenantId: string;     // UUID
  parentId?: string;    // UUID (for hierarchy)
  name: string;
  type: string;         // VARCHAR (no enum)
  createdAt: string;
}
```

### `GET /locations`
List all locations for the tenant.

### `GET /locations/{id}`
Get single location.

### `POST /locations`
Create a new location.

**Request Body:**
```json
{
  "name": "Building A",
  "parentId": "parent-uuid",
  "type": "Building"
}
```

### `PUT /locations/{id}`
Update a location.

### `DELETE /locations/{id}`
Delete a location.

---

## 7. Work Orders API (v1.1 Schema)

### Work Order Object Schema
```typescript
interface WorkOrder {
  id: string;              // UUID
  tenantId: string;        // UUID
  readableId: number;      // Human-readable sequential ID
  assetId?: string;        // UUID
  assignedUserId?: string; // UUID (renamed from assigneeId)
  status: WOStatus;
  origin: WOOrigin;
  priority: string;        // VARCHAR: Low, Medium, High, Critical
  title: string;           // Required title field
  description?: string;
  startedAt?: string;
  completedAt?: string;    // When work was completed
  createdAt: string;
}

type WOStatus = 'Requested' | 'Approved' | 'In_Progress' | 'Work_Complete' | 'Closed' | 'Cancelled';
type WOOrigin = 'Preventive_Auto' | 'Manual_Request' | 'Defect_Followup';
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

### `POST /work-orders`
Create a new work order.

**Request Body:**
```json
{
  "assetId": "asset-uuid",
  "origin": "Manual_Request",
  "priority": "High",
  "title": "Generator Maintenance",
  "description": "Engine making unusual noise"
}
```

---

## 8. Enums Reference (v1.1 - Capitalized)

### Asset Status
| Value | Description |
|-------|-------------|
| `Draft` | New, not yet active |
| `Active` | Operational |
| `Down` | Under maintenance |
| `Archived` | Decommissioned |
| `Red_Tag` | Safety concern |

### Work Order Status
| Value | Next States |
|-------|-------------|
| `Requested` | Approved, Cancelled |
| `Approved` | In_Progress, Cancelled |
| `In_Progress` | Work_Complete |
| `Work_Complete` | Closed |
| `Closed` | - |
| `Cancelled` | - |

### Work Order Origin
| Value | Description |
|-------|-------------|
| `Preventive_Auto` | Scheduled PM |
| `Manual_Request` | Corrective maintenance |
| `Defect_Followup` | From defect report |

---

## 9. Files API (MinIO Storage)

File uploads are stored in MinIO object storage, organized by tenant and asset.

### `POST /upload`
Upload a file (multipart form-data).

**Request:**
- Content-Type: `multipart/form-data`
- Max Size: **10 MB**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | File | ✓ | The file to upload |
| `assetId` | string | | Associate with asset |

**Response (201 Created):**
```json
{
  "success": true,
  "objectName": "tenants/tenant-id/assets/asset-id/1234567890.jpg",
  "size": 102400,
  "contentType": "image/jpeg",
  "originalName": "photo.jpg",
  "downloadUrl": "/api/v1/files/attachments/..."
}
```

### `GET /files/{bucket}/{objectPath}`
Download a file (redirects to presigned URL).

### `DELETE /files/{bucket}/{objectPath}`
Delete a file.

---

## 10. Admin APIs

### `GET /audit-logs`
Paginated, filterable list of audit entries (Admin only).

### `GET /tenant/settings`
Returns current tenant settings.

### `PATCH /tenant/settings`
Partial update of settings.

---

## 11. Health & System

### `GET /health`
Health check endpoint (public, no auth required).

**Response (200 OK):**
```json
{
  "status": "up",
  "message": "It's healthy",
  "timestamp": "2026-02-01T10:00:00Z"
}
```

### `GET /system/health`
Extended health check with component status (Admin only).

**Response (200 OK):**
```json
{
  "status": "up",
  "components": {
    "database": { "status": "up", "latencyMs": 3 },
    "storage": { "status": "up", "bucketExists": true }
  },
  "stats": {
    "totalUsers": 12,
    "totalAssets": 347,
    "totalWorkOrders": 89,
    "openWorkOrders": 23
  }
}
```

---

## 15. Analytics API

### `GET /analytics/dashboard`
Aggregated stats for the executive dashboard.

**Response (200 OK):**
```json
{
  "stats": {
    "totalAssets": 1245,
    "maintenanceDue": 12,
    "openFaults": 3,
    "criticalAlerts": 3,
    "openWorkOrders": 5,
    "completionRate": 94
  },
  "alerts": [
    {
      "id": "uuid",
      "title": "Asset Reported Critical Status",
      "severity": "Critical",
      "assetId": "asset-uuid",
      "assetName": "Hydraulic Press",
      "timestamp": "2023-11-20T10:00:00Z"
    }
  ]
}
```

---

## 16. User Management (Enhanced)

### `PUT /users/{id}`
Update user details (Role, Status, Name).

**Request Body:**
```json
{
  "fullName": "John Doe",
  "role": "Manager",
  "isActive": true,
  "orgUnitId": null
}
```

---

## 17. Tenant Settings (Enhanced)

### `PATCH /tenant/settings`
Update system configuration (Partial update).

**Request Body:**
```json
{
  "settings": {
    "maintenanceCycleDays": 90,
    "assetCategories": ["Hardware", "software"],
    "erpConfig": {
      "enabled": true,
      "provider": "SAP"
    }
  }
}
```

---

## 12. Parts Catalog API

### Part Object Schema
```typescript
interface Part {
  id: string;
  tenantId: string;
  sku: string;
  name: string;
  category?: string;
  uom: string;          // Unit of measure (Each, Box, Liter, etc.)
  minStockLevel: number;
  isStockItem: boolean;
  createdAt: string;
}
```

### `GET /parts`
List all parts with filtering.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `category` | string | Filter by category |
| `search` | string | Search in name/SKU |
| `page` | int | Page number |
| `limit` | int | Items per page |

### `POST /parts`
Create a new part.

**Request Body:**
```json
{
  "sku": "FLT-001",
  "name": "Oil Filter",
  "category": "Filters",
  "uom": "Each",
  "minStockLevel": 10,
  "isStockItem": true
}
```

### `PUT /parts/{id}`
Update a part.

### `DELETE /parts/{id}`
Delete a part.

---

## 13. Inventory Stock API

### InventoryStock Object Schema
```typescript
interface InventoryStock {
  id: string;
  tenantId: string;
  partId: string;
  locationId: string;
  quantityOnHand: number;
  binLabel?: string;
  updatedAt: string;
  partName: string;   // Joined
  partSku: string;    // Joined
  locationName: string; // Joined
}
```

### `GET /inventory/stock`
List inventory stock with filtering.

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `partId` | string | Filter by part |
| `locationId` | string | Filter by location |
| `lowStock` | boolean | Show only items below min level |
| `page` | int | Page number |
| `limit` | int | Items per page |

### `POST /inventory/stock`
Create or update stock at a location.

**Request Body:**
```json
{
  "partId": "part-uuid",
  "locationId": "location-uuid",
  "quantityOnHand": 50,
  "binLabel": "A-01-03"
}
```

### `POST /inventory/stock/adjust`
Adjust stock quantity (positive or negative).

**Request Body:**
```json
{
  "partId": "part-uuid",
  "locationId": "location-uuid",
  "delta": -5
}
```

---

## 14. Inventory Wallets API

Technician-held inventory (assigned stock).

### `GET /inventory/wallets`
Get current user's wallet.

### `GET /inventory/wallets/{userId}`
Get a specific user's wallet.
