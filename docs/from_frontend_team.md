Context: These requirements are derived from the frontend implementation of the IOI Asset & Maintenance Management System (AMMS). Compliance is required to ensure the UI renders correctly and features like filtering and pagination work as expected.

1. Data Models & Enums
1.1 Asset Entity
The frontend expects an 
Asset
 object with the following structure. Please ensure the API response matches these keys exactly (camelCase).

{
  "id": "string (AST-XXX)",
  "name": "string",
  "location": "string",
  "status": "enum (see below)",
  "criticality": "enum (see below)",
  "lastInspection": "ISO 8601 Date String (YYYY-MM-DD)"
}
1.2 Strict Enum Values
The frontend logic (badge coloring, filtering) relies on these exact string values. Do not deviate (e.g., do not send "In Progress" instead of "maintenance").

Status Enums:

operational (Maps to Success/Green)
maintenance (Maps to Warning/Yellow)
decommissioned (Maps to Danger/Red)
pending (Maps to Info/Blue)
Criticality Enums:

high (Maps to Danger/Red)
medium (Maps to Warning/Yellow)
low (Maps to Info/Blue)
2. API Design & Features
2.1 Pagination & Sorting
The frontend uses TanStack Table in "manual" (server-side) mode. The API must support:

Pagination: 
page
 (1-based) and limit (page size) query parameters.
Response Metadata: The response must include total counts for pagination UI.
{
  "data": [...],
  "meta": {
    "total": 150,
    "page": 1,
    "limit": 10,
    "totalPages": 15
  }
}
Sorting: sortBy (field name) and sortDir (asc | desc).
2.2 Filtering
Ideally, support filtering by multiple statuses or criticalities via query params:

GET /api/assets?status=operational&status=maintenance
GET /api/assets?criticality=high
3. Error Handling
The frontend uses Sonner for toast notifications.

4xx/5xx Errors: Return a standard error object so we can display meaningful messages.
{
  "error": "Conflict",
  "message": "Asset AST-001 cannot be decommissioned while active work orders exist."
}
4. Performance
Latency: Dashboard KPIs and Asset Lists should load in < 500ms to maintain the "premium" feel.
Search: Asset search (by name/ID) should be optimized for debounce patterns.