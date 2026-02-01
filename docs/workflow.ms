# IOI AMMS: System Architecture & Workflow Guide
**Version:** 1.1 (Final)
**Philosophy:** Disruption through Simplicity
**Architecture:** Multi-Tenant, Recursive Hierarchies, Hybrid Relational/JSONB

---

## 1. Executive Summary
IOI AMMS is designed to bridge the gap between "Financial Reality" (ERP/Depreciation) and "Operational Reality" (Maintenance/Uptime). Unlike traditional bloatware, it prioritizes:
1.  **Speed:** <30s interactions for field technicians.
2.  **Truth:** Separate Operational data from Financial data to prevent "dirty" overwrites.
3.  **Flexibility:** "Hybrid" maintenance triggers that combine Calendar dates with Usage meters.

---

## 2. Core Modules & Schema Explanation

### Module A: Identity & Organization (The Skeleton)
Before any asset is added, the corporate structure is defined. We use **Recursive Hierarchies** to support complex matrix organizations.
* **`tenants`**: The root container. Isolates data between different clients.
* **`org_units`**: A recursive tree defining *Functional Ownership*.
    * *Structure:* Region $\rightarrow$ Country $\rightarrow$ Business Unit $\rightarrow$ Business Line $\rightarrow$ Team.
    * *Purpose:* Allows reporting like "Show me maintenance spend for the Wireline Business Line across all Countries."
* **`users`**: System actors linked to a specific `org_unit_id` (their home team).

### Module B: The Living Registry (Assets & Locations)
* **`locations`**: A recursive tree defining *Physical Placement* (Site $\rightarrow$ Room $\rightarrow$ Shelf).
* **`assets`**: The operational definition of equipment.
    * *Recursive:* Assets can be parents of other assets (e.g., A "Compressor Skid" contains a "Motor").
    * *Flexible:* Uses `specs` (JSONB) to store variable technical data (Voltage, RPM) without requiring infinite database columns.
    * *Field Verified:* Includes an `is_field_verified` flag for "Walkdown" campaigns.
* **`asset_finance`**: A "Sidecar Table" for sensitive data (`acquisition_cost`, `book_value`, `erp_fixed_asset_id`). Locked down to Managers/Admins.
* **`asset_identities`**: Decouples the database UUID from `Client_Code` (P-101) and `QR_Token`.
* **`assets_staging`**: The "Mud Room" where raw imports land before being cleaned and promoted.

### Module C: The Maintenance Engine (The Brain)
* **`checklist_templates`**: Reusable definitions of tasks (JSONB) used to spawn Work Order checklists.
* **`pm_schedules`**: Defines the "Race to Zero" logic.
    * *The Logic:* A PM is due if `(Date >= Target)` **OR** `(Hours >= Target)` **OR** `(KM >= Target)`.
    * *Suppression:* Can suppress smaller PMs (L1) when a major PM (L2) triggers.
* **`work_orders`**: A Single Stream transaction table replacing separate "Requests" and "Orders."
    * *Status Flow:* `Requested` $\rightarrow$ `Approved` $\rightarrow$ `In_Progress` $\rightarrow$ `Closed`.
* **`wo_tasks`**: The individual execution steps spawned from templates. Allows SQL querying on specific checks (e.g., "Show all failed Safety Checks").
* **`wo_labor_logs` & `wo_resource_logs`**: Normalized tables for precise costing of Time and Materials.

### Module D: Logistics & Inventory (The Wallet)
* **`parts`**: The Item Master catalog (SKU, Name, Min Stock).
* **`inventory_stock`**: Physical counts in Warehouse Locations (`bin_label`).
* **`inventory_wallets`**: Tracks "Personal Stock."
    * When a Technician takes parts, they move to their "Wallet." They are not expensed until used on a Work Order.
* **`asset_movements`**: Tracks Chain of Custody (`In_Transit`, `Received`) linked to external Gate Passes.
* **`verification_campaigns`**: Project-based audit tools to snapshot and verify assets in a specific zone.

---

## 3. Core Workflows (The "How-To")

### Workflow 1: Onboarding a New Client
1.  **Define Hierarchy:** Admin sets up `org_units` and `locations`.
2.  **Import Data:** Upload Excel to `assets_staging`.
3.  **Process:** Manager maps columns. System creates `assets` and `asset_finance` records.
4.  **Field Verify:** Technicians scan QR tags to link `asset_identities` to the new records.

### Workflow 2: The Hybrid Maintenance Trigger
* **Scenario:** Generator requires service every 6 Months **OR** 500 Hours.
* **Day 90:**
    * Time Check: 3 months (Not due).
    * Meter Check: Daily Log reports 600 Hours.
    * **Result:** 600 > 500. System triggers `preventive_auto` Work Order. The "6 Month" timer is reset.

### Workflow 3: The "Technician Wallet" (Inventory)
1.  **Checkout:** Technician scans filters at Warehouse. Inventory moves: `Warehouse` $\rightarrow$ `Technician Wallet`.
2.  **Execution:** Technician fixes pump. Selects Filter from "My Wallet" in the Work Order.
3.  **Result:** Part expensed to Asset. Wallet quantity decreases.

### Workflow 4: The Defect Loop (Catching Failures)
1.  **Execution:** Technician ticks "Fail" on task "Check Engine Mounts."
2.  **System:** Detects failure in `wo_tasks`.
3.  **Action:** Instantly spawns a linked `Defect_Followup` Work Order to ensure the issue is tracked.