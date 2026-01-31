# Database Schema Overview
**Project:** IOI AMMS (MVP)
**Database:** PostgreSQL 16+
**Key Strategy:** * **Recursive Hierarchies** for Locations/Assets (Adjacency List).
* **JSONB** for flexible metadata (Specs, Form Inputs).
* **Strict Foreign Keys** for data integrity.

NOTE: THIS IS A WORK IN PROGRESS AND BY NO MEANS FINAL.
---

## 1. Core & Identity (Multi-Tenancy)
*Foundational access control and tenant isolation.*

| Table Name | Description | Key Columns |
| :--- | :--- | :--- |
| **`tenants`** | Represents the Client Organization. | `id` (UUID), `name`, `settings` (JSONB - *Toggle Agile Mode here*) |
| **`users`** | System users (Techs, Managers). | `id`, `tenant_id`, `email`, `role` (Enum: `technician`, `supervisor`...), `password_hash` |
| **`audit_logs`** | Who changed what and when. | `id`, `user_id`, `action`, `entity_type`, `entity_id`, `changes` (JSONB) |

---

## 2. The Living Registry (Assets & Locations)
*The "Source of Truth" for physical equipment.*

| Table Name | Description | Key Columns |
| :--- | :--- | :--- |
| **`locations`** | Physical places (Site -> Area). | `id`, `parent_id` (Self-ref), `name`, `type` (Site, Building, Room) |
| **`assets`** | The physical equipment & units. | `id` (UUID), `tenant_id`, `parent_id` (for Components inside Units), `location_id`, `name` |
| **`asset_identities`** | Search keys for an asset. | `asset_id`, `client_code` (e.g. P-101), `qr_token` (Secure URL ID), `barcode_manufacturer` |
| **`asset_meters`** | Current health/usage stats. | `asset_id`, `current_run_hours`, `current_odometer`, `last_updated_at` |
| **`asset_specs`** | Flexible technical data. | `asset_id`, `specs` (JSONB - e.g. `{"voltage": "220v", "rpm": 1400}`) |

---

## 3. The Maintenance Engine
*Work Order generation, execution, and history.*

| Table Name | Description | Key Columns |
| :--- | :--- | :--- |
| **`pm_schedules`** | The "Hybrid" Trigger logic. | `id`, `asset_id`, `interval_days`, `interval_meter`, `last_performed_date`, `suppressed_by_schedule_id` |
| **`work_orders`** | The central transaction. | `id`, `asset_id`, `status` (Draft, Ready, In_Progress, Closed), `origin` (PM, CM, Defect), `priority` |
| **`wo_tasks`** | Checklist items for a WO. | `id`, `wo_id`, `description`, `is_mandatory`, `result` (Pass/Fail/Input) |
| **`service_requests`** | Triage queue for Operators. | `id`, `asset_id`, `reported_by_user_id`, `description`, `photo_url`, `status` (Pending, Rejected, Converted) |

---

## 4. Resources & Inventory (Lite)
*The "Technician Wallet" and consumption tracking.*

| Table Name | Description | Key Columns |
| :--- | :--- | :--- |
| **`parts`** | Item Master (Catalog). | `id`, `sku`, `name`, `category`, `min_level` |
| **`inventory_stock`** | Quantity at a specific warehouse. | `id`, `part_id`, `location_id` (Warehouse), `qty_on_hand`, `bin_location` |
| **`inventory_wallets`** | **The "Tech Wallet".** | `id`, `user_id` (Technician), `part_id`, `qty_held` |
| **`wo_consumables`** | Parts used on a specific Job. | `id`, `wo_id`, `part_id`, `qty_used`, `cost_at_time`, `source` (Wallet vs. Warehouse) |

---

## 5. Logistics (Chain of Custody)
*Tracking movements between sites.*

| Table Name | Description | Key Columns |
| :--- | :--- | :--- |
| **`asset_movements`** | The Transfer history. | `id`, `asset_id`, `from_location_id`, `to_location_id`, `status` (In_Transit, Received), `gate_pass_ref` |
| **`shipments`** | Grouping transfers (The Truck). | `id`, `jms_reference` (External Trip ID), `driver_name`, `manifest_pdf_url` |

---

## 6. Daily Ops (The Gap Filler)
*Capturing utilization data from the field.*

| Table Name | Description | Key Columns |
| :--- | :--- | :--- |
| **`daily_logs`** | Bulk entry of hours/fuel. | `id`, `date`, `supervisor_id`, `asset_id`, `run_hours_added`, `fuel_liters_added` |
