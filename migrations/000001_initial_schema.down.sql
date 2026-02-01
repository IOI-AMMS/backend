-- Down migration for v1.1 schema
-- Drops all tables in reverse order of creation

DROP TABLE IF EXISTS audit_logs;

DROP TABLE IF EXISTS attachments;

DROP TABLE IF EXISTS verification_entries;

DROP TABLE IF EXISTS verification_campaigns;

DROP TABLE IF EXISTS asset_movements;

DROP TABLE IF EXISTS inventory_wallets;

DROP TABLE IF EXISTS inventory_stock;

DROP TABLE IF EXISTS parts;

DROP TABLE IF EXISTS wo_resource_logs;

DROP TABLE IF EXISTS wo_labor_logs;

DROP TABLE IF EXISTS wo_tasks;

DROP TABLE IF EXISTS work_orders;

DROP TABLE IF EXISTS pm_schedules;

DROP TABLE IF EXISTS checklist_templates;

DROP TABLE IF EXISTS asset_meters;

DROP TABLE IF EXISTS assets_staging;

DROP TABLE IF EXISTS asset_identities;

DROP TABLE IF EXISTS asset_finance;

DROP TABLE IF EXISTS assets;

DROP TABLE IF EXISTS locations;

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS org_units;

DROP TABLE IF EXISTS tenants;

DROP EXTENSION IF EXISTS "uuid-ossp";