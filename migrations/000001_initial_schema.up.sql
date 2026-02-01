-- IOI AMMS - Master Database Schema (Final MVP)
-- Version: 1.1
-- Database: PostgreSQL 16+
-- Migration: 000001_initial_schema

-- 0. EXTENSIONS
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==========================================
-- MODULE 1: IDENTITY & ORG STRUCTURE
-- ==========================================

-- 1. TENANTS (The Client Container)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(100) UNIQUE,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- 2. ORG UNITS (Recursive Functional Hierarchy)
CREATE TABLE org_units (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    parent_id UUID REFERENCES org_units (id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    cost_center_code VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- 3. USERS (System Actors)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    org_unit_id UUID REFERENCES org_units (id),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    role VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ==========================================
-- MODULE 2: THE LIVING REGISTRY
-- ==========================================

-- 4. LOCATIONS (Recursive Physical Hierarchy)
CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    parent_id UUID REFERENCES locations (id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 5. ASSETS (The Operational Truth)
CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    parent_id UUID REFERENCES assets (id),
    location_id UUID REFERENCES locations (id),
    org_unit_id UUID REFERENCES org_units (id),
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'Draft',
    is_field_related BOOLEAN DEFAULT TRUE,
    is_field_verified BOOLEAN DEFAULT FALSE,
    manufacturer VARCHAR(100),
    model_number VARCHAR(100),
    specs JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 6. ASSET FINANCE (Sensitive Data - 1:1 Link)
CREATE TABLE asset_finance (
    asset_id UUID PRIMARY KEY REFERENCES assets (id) ON DELETE CASCADE,
    purchase_date DATE,
    placed_in_service_date DATE,
    warranty_expiry_date DATE,
    currency VARCHAR(3) DEFAULT 'USD',
    acquisition_cost DECIMAL(15, 2),
    current_book_value DECIMAL(15, 2),
    total_depreciation DECIMAL(15, 2),
    vendor_name VARCHAR(255),
    erp_fixed_asset_id VARCHAR(100),
    erp_status VARCHAR(50),
    last_synced_at TIMESTAMP
);

-- 7. ASSET IDENTITIES (Search Keys)
CREATE TABLE asset_identities (
    asset_id UUID REFERENCES assets (id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    value VARCHAR(255) NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (asset_id, type)
);

CREATE UNIQUE INDEX idx_ident_value ON asset_identities (value);

-- 8. ASSETS STAGING (The Import Buffer)
CREATE TABLE assets_staging (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    batch_id UUID,
    import_status VARCHAR(50) DEFAULT 'Pending',
    raw_data JSONB,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ==========================================
-- MODULE 3: MAINTENANCE ENGINE
-- ==========================================

-- 9. ASSET METERS (Current Stats)
CREATE TABLE asset_meters (
    asset_id UUID PRIMARY KEY REFERENCES assets (id) ON DELETE CASCADE,
    current_run_hours DECIMAL(10, 2) DEFAULT 0,
    current_odometer_km DECIMAL(10, 2) DEFAULT 0,
    last_updated_at TIMESTAMP,
    updated_by_user_id UUID
);

-- 10. CHECKLIST TEMPLATES (Reusable Definitions)
CREATE TABLE checklist_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    items JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT NOW()
);

-- 11. PM SCHEDULES (The Hybrid Trigger)
CREATE TABLE pm_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    asset_id UUID REFERENCES assets (id),
    checklist_template_id UUID REFERENCES checklist_templates (id),
    title VARCHAR(255) NOT NULL,
    interval_days INT,
    interval_run_hours INT,
    interval_odometer_km INT,
    last_performed_date DATE,
    last_performed_run_hours DECIMAL(10, 2),
    last_performed_odometer_km DECIMAL(10, 2),
    suppress_pm_ids JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT TRUE
);

-- 12. WORK ORDERS (Single Stream Transaction)
CREATE TABLE work_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    readable_id SERIAL,
    asset_id UUID REFERENCES assets (id),
    assigned_user_id UUID REFERENCES users (id),
    status VARCHAR(50) NOT NULL DEFAULT 'Requested',
    origin VARCHAR(50) NOT NULL,
    priority VARCHAR(20) DEFAULT 'Medium',
    title VARCHAR(255) NOT NULL,
    description TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 13. WO TASKS (Execution Steps)
CREATE TABLE wo_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    work_order_id UUID REFERENCES work_orders (id) ON DELETE CASCADE,
    description VARCHAR(255) NOT NULL,
    task_type VARCHAR(50) NOT NULL,
    is_mandatory BOOLEAN DEFAULT FALSE,
    sort_order INT DEFAULT 0,
    result_value TEXT,
    result_notes TEXT,
    photo_url TEXT,
    completed_at TIMESTAMP,
    completed_by_user_id UUID REFERENCES users (id)
);

CREATE INDEX idx_wo_tasks_result ON wo_tasks (work_order_id, result_value);

-- 14. WO LOGS (Time & Materials)
CREATE TABLE wo_labor_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    work_order_id UUID REFERENCES work_orders (id),
    user_id UUID REFERENCES users (id),
    hours_spent DECIMAL(5, 2) NOT NULL,
    date_performed DATE DEFAULT CURRENT_DATE,
    comment TEXT
);

CREATE TABLE wo_resource_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    work_order_id UUID REFERENCES work_orders (id),
    part_name VARCHAR(255),
    part_id UUID,
    quantity DECIMAL(10, 2) NOT NULL,
    source VARCHAR(50) DEFAULT 'Warehouse'
);

-- ==========================================
-- MODULE 4: LOGISTICS & AUDIT
-- ==========================================

-- 15. PARTS CATALOG (Item Master)
CREATE TABLE parts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(50),
    uom VARCHAR(20) DEFAULT 'Each',
    min_stock_level DECIMAL(10, 2) DEFAULT 0,
    is_stock_item BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_parts_sku ON parts (tenant_id, sku);

-- 16. INVENTORY STOCK (Warehouse Quantities)
CREATE TABLE inventory_stock (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    part_id UUID REFERENCES parts (id),
    location_id UUID REFERENCES locations (id),
    quantity_on_hand DECIMAL(10, 2) DEFAULT 0,
    bin_label VARCHAR(50),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 17. INVENTORY WALLETS (Technician Stock)
CREATE TABLE inventory_wallets (
    user_id UUID REFERENCES users (id),
    part_id UUID REFERENCES parts (id),
    qty_held DECIMAL(10, 2) DEFAULT 0,
    last_updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, part_id)
);

-- 18. ASSET MOVEMENTS (Chain of Custody)
CREATE TABLE asset_movements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    asset_id UUID REFERENCES assets (id),
    from_location_id UUID REFERENCES locations (id),
    to_location_id UUID REFERENCES locations (id),
    status VARCHAR(50) DEFAULT 'In_Transit',
    transfer_type VARCHAR(50),
    shipment_ref VARCHAR(100),
    created_by_user_id UUID,
    received_by_user_id UUID,
    created_at TIMESTAMP DEFAULT NOW(),
    received_at TIMESTAMP
);

-- 19. VERIFICATION CAMPAIGNS (Audits)
CREATE TABLE verification_campaigns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    title VARCHAR(255),
    status VARCHAR(50) DEFAULT 'Active',
    target_org_unit_id UUID,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE verification_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    campaign_id UUID REFERENCES verification_campaigns (id),
    asset_id UUID REFERENCES assets (id),
    status VARCHAR(50) DEFAULT 'Pending',
    verified_at TIMESTAMP,
    gps_lat DECIMAL(10, 8),
    gps_long DECIMAL(11, 8)
);

-- 20. ATTACHMENTS (Polymorphic)
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    parent_type VARCHAR(50) NOT NULL,
    parent_id UUID NOT NULL,
    file_url TEXT NOT NULL,
    file_type VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- 21. AUDIT LOGS (System Tracking)
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    tenant_id UUID REFERENCES tenants (id),
    user_id UUID REFERENCES users (id),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    changes JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ==========================================
-- INDEXES
-- ==========================================
CREATE INDEX idx_users_email ON users (email);

CREATE INDEX idx_users_tenant ON users (tenant_id);

CREATE INDEX idx_assets_tenant ON assets (tenant_id);

CREATE INDEX idx_assets_status ON assets (status);

CREATE INDEX idx_assets_location ON assets (location_id);

CREATE INDEX idx_assets_org_unit ON assets (org_unit_id);

CREATE INDEX idx_work_orders_tenant ON work_orders (tenant_id);

CREATE INDEX idx_work_orders_status ON work_orders (status);

CREATE INDEX idx_work_orders_asset ON work_orders (asset_id);

CREATE INDEX idx_audit_logs_entity ON audit_logs (entity_type, entity_id);

-- ==========================================
-- SEED DATA
-- ==========================================
INSERT INTO
    tenants (id, name, subdomain, settings)
VALUES (
        '00000000-0000-0000-0000-000000000001',
        'IOI Demo',
        'ioi-demo',
        '{"agile_mode": true}'
    );

INSERT INTO
    users (
        id,
        tenant_id,
        email,
        password_hash,
        full_name,
        role
    )
VALUES (
        '00000000-0000-0000-0000-000000000002',
        '00000000-0000-0000-0000-000000000001',
        'admin@ioi.com',
        '$2a$10$5qmcTxHBdFSqso7/Pazx/uVM1eO8OPMwyQnuCqzstHjAJ7FV2urOS',
        'Admin User',
        'Admin'
    );