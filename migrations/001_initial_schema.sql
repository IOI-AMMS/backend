-- +goose Up
-- +goose StatementBegin

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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID REFERENCES users (id),
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    changes JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Living Registry (Locations & Assets)
CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    tenant_id UUID REFERENCES tenants (id),
    parent_id UUID REFERENCES locations (id),
    name TEXT NOT NULL,
    type location_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    tenant_id UUID REFERENCES tenants (id),
    location_id UUID REFERENCES locations (id),
    name TEXT NOT NULL,
    client_code TEXT,
    status asset_status NOT NULL DEFAULT 'pending',
    criticality asset_criticality NOT NULL DEFAULT 'medium',
    model TEXT,
    serial_number TEXT,
    manufacturer TEXT,
    purchase_date DATE,
    specifications JSONB DEFAULT '{}'::jsonb,
    last_inspection TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Maintenance Engine
CREATE TABLE pm_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    asset_id UUID REFERENCES assets (id),
    name TEXT NOT NULL,
    interval_days INT,
    interval_hours INT,
    last_performed TIMESTAMP WITH TIME ZONE,
    next_due TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE work_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    tenant_id UUID REFERENCES tenants (id),
    asset_id UUID REFERENCES assets (id),
    assignee_id UUID REFERENCES users (id),
    status wo_status NOT NULL DEFAULT 'Draft',
    priority wo_priority NOT NULL DEFAULT 'Medium',
    origin wo_origin NOT NULL,
    description TEXT,
    due_date TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE wo_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    work_order_id UUID REFERENCES work_orders (id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE service_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    tenant_id UUID REFERENCES tenants (id),
    asset_id UUID REFERENCES assets (id),
    requester_id UUID REFERENCES users (id),
    status sr_status NOT NULL DEFAULT 'Pending',
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Inventory & Logistics
CREATE TABLE parts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    tenant_id UUID REFERENCES tenants (id),
    name TEXT NOT NULL,
    sku TEXT,
    min_stock INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE inventory_stock (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    part_id UUID REFERENCES parts (id),
    location_id UUID REFERENCES locations (id),
    quantity INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_users_email ON users (email);

CREATE INDEX idx_users_tenant ON users (tenant_id);

CREATE INDEX idx_assets_tenant ON assets (tenant_id);

CREATE INDEX idx_assets_status ON assets (status);

CREATE INDEX idx_assets_location ON assets (location_id);

CREATE INDEX idx_work_orders_tenant ON work_orders (tenant_id);

CREATE INDEX idx_work_orders_status ON work_orders (status);

CREATE INDEX idx_work_orders_assignee ON work_orders (assignee_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inventory_stock;

DROP TABLE IF EXISTS parts;

DROP TABLE IF EXISTS service_requests;

DROP TABLE IF EXISTS wo_tasks;

DROP TABLE IF EXISTS work_orders;

DROP TABLE IF EXISTS pm_schedules;

DROP TABLE IF EXISTS assets;

DROP TABLE IF EXISTS locations;

DROP TABLE IF EXISTS audit_logs;

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS tenants;

DROP TYPE IF EXISTS transfer_status;

DROP TYPE IF EXISTS sr_status;

DROP TYPE IF EXISTS wo_priority;

DROP TYPE IF EXISTS wo_origin;

DROP TYPE IF EXISTS wo_status;

DROP TYPE IF EXISTS asset_criticality;

DROP TYPE IF EXISTS asset_status;

DROP TYPE IF EXISTS location_type;

DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd