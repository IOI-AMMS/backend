-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enums
CREATE TYPE user_role AS ENUM ('technician', 'supervisor', 'storeman', 'manager', 'admin');

CREATE TYPE location_type AS ENUM ('Site', 'Building', 'Room', 'Zone');

CREATE TYPE asset_status AS ENUM ('operational', 'maintenance', 'decommissioned', 'pending');

CREATE TYPE asset_criticality AS ENUM ('high', 'medium', 'low');

CREATE TYPE wo_status AS ENUM ('Draft', 'Ready', 'In_Progress', 'Closed');

CREATE TYPE wo_origin AS ENUM ('PM', 'CM', 'Defect');

CREATE TYPE wo_priority AS ENUM ('Low', 'Medium', 'High', 'Critical');

CREATE TYPE sr_status AS ENUM ('Pending', 'Rejected', 'Converted');

CREATE TYPE transfer_status AS ENUM ('In_Transit', 'Received', 'Disputed');

-- 1. Core & Identity
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    settings JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    tenant_id UUID REFERENCES tenants (id),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'technician',
    first_name TEXT,
    last_name TEXT,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID REFERENCES users (id),
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    changes JSONB,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

-- 2. The Living Registry
CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    tenant_id UUID REFERENCES tenants (id),
    parent_id UUID REFERENCES locations (id),
    name TEXT NOT NULL,
    type location_type NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    tenant_id UUID REFERENCES tenants (id),
    parent_id UUID REFERENCES assets (id), -- For components inside units
    location_id UUID REFERENCES locations (id),
    name TEXT NOT NULL,
    status asset_status NOT NULL DEFAULT 'pending',
    criticality asset_criticality NOT NULL DEFAULT 'medium',
    last_inspection TIMESTAMP
    WITH
        TIME ZONE,
        created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

CREATE TABLE asset_identities (
    asset_id UUID PRIMARY KEY REFERENCES assets (id) ON DELETE CASCADE,
    client_code TEXT,
    qr_token TEXT UNIQUE,
    barcode_manufacturer TEXT,
    updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

CREATE TABLE asset_meters (
    asset_id UUID PRIMARY KEY REFERENCES assets (id) ON DELETE CASCADE,
    current_run_hours NUMERIC DEFAULT 0,
    current_odometer NUMERIC DEFAULT 0,
    last_updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

CREATE TABLE asset_specs (
    asset_id UUID PRIMARY KEY REFERENCES assets(id) ON DELETE CASCADE,
    specs JSONB DEFAULT '{}'::jsonb
);

-- 3. Maintenance Engine
CREATE TABLE pm_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    asset_id UUID REFERENCES assets (id),
    interval_days INTEGER,
    interval_meter NUMERIC,
    last_performed_date TIMESTAMP
    WITH
        TIME ZONE,
        suppressed_by_schedule_id UUID REFERENCES pm_schedules (id),
        created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

CREATE TABLE work_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    tenant_id UUID REFERENCES tenants (id),
    asset_id UUID REFERENCES assets (id),
    status wo_status NOT NULL DEFAULT 'Draft',
    origin wo_origin NOT NULL,
    priority wo_priority NOT NULL DEFAULT 'Medium',
    description TEXT,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

CREATE TABLE wo_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    wo_id UUID REFERENCES work_orders (id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    is_mandatory BOOLEAN DEFAULT FALSE,
    result TEXT,
    completed_at TIMESTAMP
    WITH
        TIME ZONE
);

CREATE TABLE service_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    asset_id UUID REFERENCES assets (id),
    reported_by_user_id UUID REFERENCES users (id),
    description TEXT NOT NULL,
    photo_url TEXT,
    status sr_status NOT NULL DEFAULT 'Pending',
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

-- 4. Resources & Inventory
CREATE TABLE parts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    tenant_id UUID REFERENCES tenants (id),
    sku TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT,
    min_level INTEGER DEFAULT 0,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

CREATE TABLE inventory_stock (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    part_id UUID REFERENCES parts (id),
    location_id UUID REFERENCES locations (id), -- Warehouse
    qty_on_hand INTEGER DEFAULT 0,
    bin_location TEXT,
    updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

CREATE TABLE inventory_wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID REFERENCES users (id),
    part_id UUID REFERENCES parts (id),
    qty_held INTEGER DEFAULT 0,
    updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

CREATE TABLE wo_consumables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    wo_id UUID REFERENCES work_orders (id) ON DELETE CASCADE,
    part_id UUID REFERENCES parts (id),
    qty_used INTEGER NOT NULL,
    cost_at_time NUMERIC,
    source TEXT -- Wallet vs Warehouse
);

-- 5. Logistics
CREATE TABLE asset_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    asset_id UUID REFERENCES assets (id),
    from_location_id UUID REFERENCES locations (id),
    to_location_id UUID REFERENCES locations (id),
    status transfer_status NOT NULL DEFAULT 'In_Transit',
    gate_pass_ref TEXT,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

CREATE TABLE shipments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    jms_reference TEXT,
    driver_name TEXT,
    manifest_pdf_url TEXT,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);

-- 6. Daily Ops
CREATE TABLE daily_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    date DATE NOT NULL,
    supervisor_id UUID REFERENCES users (id),
    asset_id UUID REFERENCES assets (id),
    run_hours_added NUMERIC,
    fuel_liters_added NUMERIC,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT NOW()
);