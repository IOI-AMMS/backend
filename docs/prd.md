# Product Requirements Document (PRD)
**Product:** IOI Asset & Maintenance Management System (AMMS)
**Version:** 3.0 (Post-RBAC)
**Philosophy:** Disruption through Simplicity
**Status:** Approved for Implementation

---

## 1. Product Vision
To build a "Living Registry" that bridges the gap between Financial ERPs and Operational Reality. The system prioritizes **data verification**, **mobility**, and **speed (<30s interactions)** over enterprise complexity.
* **Core Rule:** "ERP is the Master of Existence; AMMS is the Master of Condition."

---

## 2. Technical Foundation
* **Architecture:** Single-Container Deployment (Go Binary + Embedded Assets).
* **Database:** PostgreSQL (Relational + JSONB for attributes).
* **Infrastructure:** Docker Compose (Zero external dependencies).
* **Performance:** <200ms API response; Support for 40+ concurrent users on a single VPS.
* **Connectivity:** Offline-First capability for Mobile App (Data sync when online).

---

## 3. Functional Modules (The MVP Scope)

### Module A: The Living Registry (Assets & Units)
* **Objective:** Establish the "Ground Truth" of what actually exists on site.
* **Key Features:**
    * **Hierarchy:** Strict distinction between **Units** (Functional Parents) and **Assets** (Physical Children).
    * **Verification Mode:** Mobile workflow to validate "Ghost Assets" from ERP imports.
    * **Identity Management:** Internal UUIDs vs. Editable Client Codes. QR Codes support `Scan-to-Resolve`.
    * **Sync Strategy:** One-way daily import from ERP. AMMS is master of condition/metadata.

### Module B: Logistics & Chain of Custody
* **Objective:** Track equipment moving between Rigs, Warehouses, and Workshops.
* **Key Features:**
    * **Digital Gate Pass:** PDF Manifests with "Master QR" for bulk receiving.
    * **Hybrid Transfer:** Link transfers to external JMS IDs.
    * **Lifecycle:** `Active` -> `In Transit` -> `Received` -> `Disputed`.
    * **Cascade Moves:** Moving a Parent Unit automatically moves all Child Assets.

### Module C: The Maintenance Engine (The Brain)
* **Objective:** Move from Reactive to Predictive using Hybrid Logic.
* **Key Features:**
    * **Hybrid Triggers:** "Race to Zero" logic (Date vs. Usage).
    * **Suppression:** L2 Service suppresses L1 Service.
    * **The 4 WO Origins:** Preventive, Corrective, Defect (Auto-linked), Service Request (Triage).

### Module D: Field Execution & Resources
* **Objective:** Empower the technician with "Wallet" tools and simplified workflows.
* **Key Features:**
    * **Technician Wallet:** "Check-out / Check-in" inventory model.
    * **Daily Ops Log:** Supervisor interface for bulk Run Hours/Fuel entry.
    * **Smart Close-out:** Mandatory "Failure Codes" and "Time Spent".

### Module E: Actionable Intelligence
* **Objective:** Zero Vanity Metrics.
* **Key Features:**
    * **Real-Time Dashboard:** Critical Down, Overdue PMs, Wallet Watch.
    * **Bad Actor Report:** Top 5 Assets consuming budget/time.
    * **Monday Morning Brief:** Auto-generated PDF email summary.

---

## 4. User Roles & Permissions (RBAC)

**The "Agile Toggle":** A tenant-level setting can elevate Technicians to "Agile Mode" (Self-Approval), collapsing the Supervisor approval steps for small teams.

| Feature Area | **Technician** | **Supervisor** | **Storeman** | **Ops Manager** |
| :--- | :--- | :--- | :--- | :--- |
| **Asset Registry** | View / Search / Scan | **Verify / Edit / Transfer** | View / Scan | **Archive / Delete** |
| **Asset Creation** | Create Draft (Unverified) | Create & Verify | View Only | Full Access |
| **Work Orders** | Execute / Request / Close (My Jobs) | **Approve / Assign / QC** | View Only | Full Access |
| **Inventory** | **My Wallet (Use/Return)** | Audit / Check | **Adjust / Receive** | Full Access |
| **Logistics** | Request Transfer | Approve Transfer | **Gate Pass / Receive** | Full Access |
| **Reports** | None | Team Status | Stock Reports | Financial / Strategy |

---

## 5. Success Metrics (KPIs)
1.  **Verification Velocity:** % of "Ghost Assets" verified within first 30 days.
2.  **Adoption:** Daily Active Users (DAU) vs. Total Licenses.
3.  **Data Health:** % of Assets with valid Meter Readings (updated in last 7 days).
4.  **Reaction Time:** MTTR (Mean Time To Repair).

---