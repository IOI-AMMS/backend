-- Insert default tenant
INSERT INTO
    tenants (id, name, settings)
VALUES (
        '00000000-0000-0000-0000-000000000001',
        'IOI Default Tenant',
        '{"agileMode": false}'
    );

-- Insert default admin user (password: password123)
-- Hash generated with bcrypt cost 10
INSERT INTO
    users (
        id,
        tenant_id,
        email,
        password_hash,
        role,
        first_name,
        last_name
    )
VALUES (
        '00000000-0000-0000-0000-000000000002',
        '00000000-0000-0000-0000-000000000001',
        'admin@ioi.com',
        '$2a$10$rDkQ.zS1F8S5X5P5Y5Y5YeY5Y5Y5Y5Y5Y5Y5Y5Y5Y5Y5Y5Y5Y5Y5Y',
        'manager',
        'Admin',
        'User'
    );