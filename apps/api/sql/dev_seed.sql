BEGIN;

INSERT INTO core.organizations (id, code, name, status)
VALUES ('00000000-0000-4000-8000-000000000001', 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now();

INSERT INTO core.roles (id, org_id, code, name, description, status, is_system)
VALUES
  (
    '00000000-0000-4000-8000-000000000201',
    '00000000-0000-4000-8000-000000000001',
    'ERP_ADMIN',
    'ERP Admin',
    'Local administrator role',
    'active',
    true
  ),
  (
    '00000000-0000-4000-8000-000000000202',
    '00000000-0000-4000-8000-000000000001',
    'WAREHOUSE_STAFF',
    'Warehouse Staff',
    'Local warehouse operator role',
    'active',
    true
  ),
  (
    '00000000-0000-4000-8000-000000000203',
    '00000000-0000-4000-8000-000000000001',
    'SALES_OPS',
    'Sales Ops',
    'Local sales operations role',
    'active',
    true
  )
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    is_system = EXCLUDED.is_system,
    updated_at = now();

INSERT INTO core.users (id, org_id, email, username, display_name, password_hash, status)
SELECT *
FROM (
  VALUES
    (
      '00000000-0000-4000-8000-000000000101'::uuid,
      '00000000-0000-4000-8000-000000000001'::uuid,
      'admin@example.local',
      'admin',
      'Local Admin',
      'local-only-not-for-auth',
      'active'
    ),
    (
      '00000000-0000-4000-8000-000000000102'::uuid,
      '00000000-0000-4000-8000-000000000001'::uuid,
      'warehouse_user@example.local',
      'warehouse_user',
      'Warehouse User',
      'local-only-not-for-auth',
      'active'
    ),
    (
      '00000000-0000-4000-8000-000000000103'::uuid,
      '00000000-0000-4000-8000-000000000001'::uuid,
      'sales_user@example.local',
      'sales_user',
      'Sales User',
      'local-only-not-for-auth',
      'active'
    )
) AS seed(id, org_id, email, username, display_name, password_hash, status)
WHERE NOT EXISTS (
  SELECT 1
  FROM core.users existing
  WHERE existing.org_id = seed.org_id
    AND lower(existing.email) = lower(seed.email)
);

UPDATE core.users
SET username = seed.username,
    display_name = seed.display_name,
    password_hash = seed.password_hash,
    status = seed.status,
    updated_at = now()
FROM (
  VALUES
    ('admin@example.local', 'admin', 'Local Admin', 'local-only-not-for-auth', 'active'),
    ('warehouse_user@example.local', 'warehouse_user', 'Warehouse User', 'local-only-not-for-auth', 'active'),
    ('sales_user@example.local', 'sales_user', 'Sales User', 'local-only-not-for-auth', 'active')
) AS seed(email, username, display_name, password_hash, status)
WHERE core.users.org_id = '00000000-0000-4000-8000-000000000001'
  AND lower(core.users.email) = lower(seed.email);

INSERT INTO core.permissions (id, code, module, resource, action, description)
VALUES
  ('00000000-0000-4000-8000-000000000301', 'dashboard:view', 'dashboard', 'dashboard', 'view', 'View dashboard'),
  ('00000000-0000-4000-8000-000000000302', 'warehouse:view', 'warehouse', 'warehouse', 'view', 'View warehouse'),
  ('00000000-0000-4000-8000-000000000303', 'inventory:view', 'inventory', 'inventory', 'view', 'View inventory'),
  ('00000000-0000-4000-8000-000000000304', 'sales:view', 'sales', 'sales', 'view', 'View sales'),
  ('00000000-0000-4000-8000-000000000305', 'shipping:view', 'shipping', 'shipping', 'view', 'View shipping'),
  ('00000000-0000-4000-8000-000000000306', 'returns:view', 'returns', 'returns', 'view', 'View returns'),
  ('00000000-0000-4000-8000-000000000307', 'audit-log:view', 'audit', 'audit-log', 'view', 'View audit logs'),
  ('00000000-0000-4000-8000-000000000308', 'reports:view', 'reports', 'reports', 'view', 'View reports'),
  ('00000000-0000-4000-8000-000000000309', 'settings:view', 'settings', 'settings', 'view', 'View settings'),
  ('00000000-0000-4000-8000-000000000310', 'record:create', 'shared', 'record', 'create', 'Create records'),
  ('00000000-0000-4000-8000-000000000311', 'record:export', 'shared', 'record', 'export', 'Export records')
ON CONFLICT (code) DO UPDATE
SET module = EXCLUDED.module,
    resource = EXCLUDED.resource,
    action = EXCLUDED.action,
    description = EXCLUDED.description;

INSERT INTO core.role_permissions (role_id, permission_id)
SELECT role.id, permission.id
FROM core.roles role
JOIN core.permissions permission ON permission.code IN (
  'dashboard:view',
  'warehouse:view',
  'inventory:view',
  'sales:view',
  'shipping:view',
  'returns:view',
  'audit-log:view',
  'reports:view',
  'settings:view',
  'record:create',
  'record:export'
)
WHERE role.org_id = '00000000-0000-4000-8000-000000000001'
  AND role.code = 'ERP_ADMIN'
ON CONFLICT DO NOTHING;

INSERT INTO core.role_permissions (role_id, permission_id)
SELECT role.id, permission.id
FROM core.roles role
JOIN core.permissions permission ON permission.code IN (
  'dashboard:view',
  'warehouse:view',
  'inventory:view',
  'shipping:view',
  'returns:view'
)
WHERE role.org_id = '00000000-0000-4000-8000-000000000001'
  AND role.code = 'WAREHOUSE_STAFF'
ON CONFLICT DO NOTHING;

INSERT INTO core.role_permissions (role_id, permission_id)
SELECT role.id, permission.id
FROM core.roles role
JOIN core.permissions permission ON permission.code IN (
  'dashboard:view',
  'sales:view',
  'shipping:view',
  'returns:view',
  'reports:view',
  'record:create'
)
WHERE role.org_id = '00000000-0000-4000-8000-000000000001'
  AND role.code = 'SALES_OPS'
ON CONFLICT DO NOTHING;

INSERT INTO core.user_roles (user_id, role_id, scope_type, assigned_by)
SELECT seed.user_id, seed.role_id, seed.scope_type, seed.assigned_by
FROM (
  VALUES
    (
      '00000000-0000-4000-8000-000000000101'::uuid,
      '00000000-0000-4000-8000-000000000201'::uuid,
      'company',
      '00000000-0000-4000-8000-000000000101'::uuid
    ),
    (
      '00000000-0000-4000-8000-000000000102'::uuid,
      '00000000-0000-4000-8000-000000000202'::uuid,
      'company',
      '00000000-0000-4000-8000-000000000101'::uuid
    ),
    (
      '00000000-0000-4000-8000-000000000103'::uuid,
      '00000000-0000-4000-8000-000000000203'::uuid,
      'company',
      '00000000-0000-4000-8000-000000000101'::uuid
    )
) AS seed(user_id, role_id, scope_type, assigned_by)
WHERE NOT EXISTS (
  SELECT 1
  FROM core.user_roles existing
  WHERE existing.user_id = seed.user_id
    AND existing.role_id = seed.role_id
    AND existing.scope_type = seed.scope_type
    AND existing.scope_id IS NULL
);

INSERT INTO mdm.units (id, org_id, code, name, precision_scale, status, created_by, updated_by)
VALUES
  (
    '00000000-0000-4000-8000-000000000401',
    '00000000-0000-4000-8000-000000000001',
    'PCS',
    'Piece',
    0,
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000000402',
    '00000000-0000-4000-8000-000000000001',
    'BOX',
    'Box',
    0,
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  )
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    precision_scale = EXCLUDED.precision_scale,
    status = EXCLUDED.status,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO mdm.suppliers (id, org_id, code, name, supplier_type, email, phone, status, created_by, updated_by)
VALUES (
  '00000000-0000-4000-8000-000000000501',
  '00000000-0000-4000-8000-000000000001',
  'SUP-LOCAL-001',
  'Local Cosmetics Supplier',
  'supplier',
  'supplier@example.local',
  '0900000001',
  'active',
  '00000000-0000-4000-8000-000000000101',
  '00000000-0000-4000-8000-000000000101'
)
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    supplier_type = EXCLUDED.supplier_type,
    email = EXCLUDED.email,
    phone = EXCLUDED.phone,
    status = EXCLUDED.status,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO mdm.customers (id, org_id, code, name, email, phone, status, created_by, updated_by)
VALUES (
  '00000000-0000-4000-8000-000000000601',
  '00000000-0000-4000-8000-000000000001',
  'CUS-LOCAL-001',
  'Local Retail Customer',
  'customer@example.local',
  '0900000002',
  'active',
  '00000000-0000-4000-8000-000000000101',
  '00000000-0000-4000-8000-000000000101'
)
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    email = EXCLUDED.email,
    phone = EXCLUDED.phone,
    status = EXCLUDED.status,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO mdm.carriers (id, org_id, code, name, contact_name, phone, status, created_by, updated_by)
VALUES (
  '00000000-0000-4000-8000-000000000701',
  '00000000-0000-4000-8000-000000000001',
  'CAR-LOCAL-001',
  'Local Delivery Carrier',
  'Local Dispatcher',
  '0900000003',
  'active',
  '00000000-0000-4000-8000-000000000101',
  '00000000-0000-4000-8000-000000000101'
)
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    contact_name = EXCLUDED.contact_name,
    phone = EXCLUDED.phone,
    status = EXCLUDED.status,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO mdm.warehouses (id, org_id, code, name, address, status, created_by, updated_by)
VALUES
  (
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000000001',
    'warehouse_main',
    'Main Warehouse',
    'Local main warehouse',
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000000802',
    '00000000-0000-4000-8000-000000000001',
    'warehouse_return',
    'Return Warehouse',
    'Local return warehouse',
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  )
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    address = EXCLUDED.address,
    status = EXCLUDED.status,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO mdm.warehouse_zones (id, org_id, warehouse_id, code, name, zone_type, status, created_by, updated_by)
VALUES
  (
    '00000000-0000-4000-8000-000000000901',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000000801',
    'MAIN-STORAGE',
    'Main Storage',
    'storage',
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000000902',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000000802',
    'RETURN-STAGING',
    'Return Staging',
    'return',
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  )
ON CONFLICT (warehouse_id, code) DO UPDATE
SET name = EXCLUDED.name,
    zone_type = EXCLUDED.zone_type,
    status = EXCLUDED.status,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO mdm.warehouse_bins (id, org_id, warehouse_id, zone_id, code, name, bin_type, status, created_by, updated_by)
VALUES
  (
    '00000000-0000-4000-8000-000000001001',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000000901',
    'A-01',
    'Main Bin A-01',
    'storage',
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001002',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000000802',
    '00000000-0000-4000-8000-000000000902',
    'R-01',
    'Return Bin R-01',
    'return',
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  )
ON CONFLICT (warehouse_id, code) DO UPDATE
SET zone_id = EXCLUDED.zone_id,
    name = EXCLUDED.name,
    bin_type = EXCLUDED.bin_type,
    status = EXCLUDED.status,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO mdm.items (
  id,
  org_id,
  sku,
  name,
  item_type,
  base_unit_id,
  barcode,
  shelf_life_days,
  requires_batch,
  requires_expiry,
  status,
  created_by,
  updated_by
)
VALUES
  (
    '00000000-0000-4000-8000-000000001101',
    '00000000-0000-4000-8000-000000000001',
    'FG-LIP-001',
    'Matte Lipstick',
    'finished_good',
    '00000000-0000-4000-8000-000000000401',
    '893000000001',
    730,
    true,
    true,
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001102',
    '00000000-0000-4000-8000-000000000001',
    'FG-SER-001',
    'Vitamin C Serum',
    'finished_good',
    '00000000-0000-4000-8000-000000000401',
    '893000000002',
    540,
    true,
    true,
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001103',
    '00000000-0000-4000-8000-000000000001',
    'FG-CRM-001',
    'Moisturizing Cream',
    'finished_good',
    '00000000-0000-4000-8000-000000000401',
    '893000000003',
    730,
    true,
    true,
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001104',
    '00000000-0000-4000-8000-000000000001',
    'FG-SUN-001',
    'Daily Sunscreen',
    'finished_good',
    '00000000-0000-4000-8000-000000000401',
    '893000000004',
    540,
    true,
    true,
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001105',
    '00000000-0000-4000-8000-000000000001',
    'PKG-BOX-001',
    'Gift Box',
    'packaging',
    '00000000-0000-4000-8000-000000000402',
    '893000000005',
    NULL,
    true,
    false,
    'active',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  )
ON CONFLICT (org_id, sku) DO UPDATE
SET name = EXCLUDED.name,
    item_type = EXCLUDED.item_type,
    base_unit_id = EXCLUDED.base_unit_id,
    barcode = EXCLUDED.barcode,
    shelf_life_days = EXCLUDED.shelf_life_days,
    requires_batch = EXCLUDED.requires_batch,
    requires_expiry = EXCLUDED.requires_expiry,
    status = EXCLUDED.status,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO inventory.batches (
  id,
  org_id,
  item_id,
  batch_no,
  supplier_id,
  mfg_date,
  expiry_date,
  qc_status,
  status,
  batch_ref,
  org_ref,
  item_ref,
  supplier_ref,
  created_by_ref,
  updated_by_ref,
  created_by,
  updated_by
)
VALUES
  (
    '00000000-0000-4000-8000-000000001201',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001101',
    'BATCH-LIP-LOCAL',
    '00000000-0000-4000-8000-000000000501',
    '2026-01-01',
    '2028-01-01',
    'pass',
    'active',
    '00000000-0000-4000-8000-000000001201',
    'org-my-pham',
    'FG-LIP-001',
    'sup-local',
    'user-erp-admin',
    'user-erp-admin',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001202',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001102',
    'BATCH-SER-LOCAL',
    '00000000-0000-4000-8000-000000000501',
    '2026-01-01',
    '2027-07-01',
    'pass',
    'active',
    '00000000-0000-4000-8000-000000001202',
    'org-my-pham',
    'FG-SER-001',
    'sup-local',
    'user-erp-admin',
    'user-erp-admin',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001203',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001103',
    'BATCH-CRM-LOCAL',
    '00000000-0000-4000-8000-000000000501',
    '2026-01-01',
    '2028-01-01',
    'pass',
    'active',
    '00000000-0000-4000-8000-000000001203',
    'org-my-pham',
    'FG-CRM-001',
    'sup-local',
    'user-erp-admin',
    'user-erp-admin',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001204',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001104',
    'BATCH-SUN-LOCAL',
    '00000000-0000-4000-8000-000000000501',
    '2026-01-01',
    '2027-07-01',
    'pass',
    'active',
    '00000000-0000-4000-8000-000000001204',
    'org-my-pham',
    'FG-SUN-001',
    'sup-local',
    'user-erp-admin',
    'user-erp-admin',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001205',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001105',
    'BATCH-BOX-LOCAL',
    '00000000-0000-4000-8000-000000000501',
    '2026-01-01',
    NULL,
    'pass',
    'active',
    '00000000-0000-4000-8000-000000001205',
    'org-my-pham',
    'PKG-BOX-001',
    'sup-local',
    'user-erp-admin',
    'user-erp-admin',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001206',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001102',
    'LOT-2604A',
    '00000000-0000-4000-8000-000000000501',
    '2026-04-01',
    '2027-04-01',
    'hold',
    'active',
    'batch-serum-2604a',
    'org-my-pham',
    'item-serum-30ml',
    'sup-rm-bioactive',
    'user-erp-admin',
    'user-erp-admin',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  )
ON CONFLICT (item_id, batch_no) DO UPDATE
SET supplier_id = EXCLUDED.supplier_id,
    mfg_date = EXCLUDED.mfg_date,
    expiry_date = EXCLUDED.expiry_date,
    qc_status = EXCLUDED.qc_status,
    status = EXCLUDED.status,
    batch_ref = EXCLUDED.batch_ref,
    org_ref = EXCLUDED.org_ref,
    item_ref = EXCLUDED.item_ref,
    supplier_ref = EXCLUDED.supplier_ref,
    created_by_ref = EXCLUDED.created_by_ref,
    updated_by_ref = EXCLUDED.updated_by_ref,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO inventory.goods_receipts (
  id,
  org_id,
  grn_no,
  warehouse_id,
  supplier_id,
  receipt_date,
  status,
  created_by,
  updated_by,
  submitted_at,
  submitted_by,
  approved_at,
  approved_by
)
VALUES (
  '00000000-0000-4000-8000-000000001301',
  '00000000-0000-4000-8000-000000000001',
  'GRN-LOCAL-0001',
  '00000000-0000-4000-8000-000000000801',
  '00000000-0000-4000-8000-000000000501',
  '2026-04-26',
  'received',
  '00000000-0000-4000-8000-000000000101',
  '00000000-0000-4000-8000-000000000101',
  now(),
  '00000000-0000-4000-8000-000000000101',
  now(),
  '00000000-0000-4000-8000-000000000101'
)
ON CONFLICT (org_id, grn_no) DO UPDATE
SET warehouse_id = EXCLUDED.warehouse_id,
    supplier_id = EXCLUDED.supplier_id,
    receipt_date = EXCLUDED.receipt_date,
    status = EXCLUDED.status,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by,
    submitted_at = EXCLUDED.submitted_at,
    submitted_by = EXCLUDED.submitted_by,
    approved_at = EXCLUDED.approved_at,
    approved_by = EXCLUDED.approved_by;

INSERT INTO inventory.goods_receipt_lines (
  id,
  org_id,
  goods_receipt_id,
  line_no,
  item_id,
  batch_id,
  unit_id,
  received_qty,
  accepted_qty,
  rejected_qty,
  created_by,
  updated_by
)
VALUES
  (
    '00000000-0000-4000-8000-000000001311',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001301',
    1,
    '00000000-0000-4000-8000-000000001101',
    '00000000-0000-4000-8000-000000001201',
    '00000000-0000-4000-8000-000000000401',
    120,
    120,
    0,
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001312',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001301',
    2,
    '00000000-0000-4000-8000-000000001102',
    '00000000-0000-4000-8000-000000001202',
    '00000000-0000-4000-8000-000000000401',
    80,
    80,
    0,
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001313',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001301',
    3,
    '00000000-0000-4000-8000-000000001103',
    '00000000-0000-4000-8000-000000001203',
    '00000000-0000-4000-8000-000000000401',
    60,
    60,
    0,
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001314',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001301',
    4,
    '00000000-0000-4000-8000-000000001104',
    '00000000-0000-4000-8000-000000001204',
    '00000000-0000-4000-8000-000000000401',
    90,
    90,
    0,
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001315',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001301',
    5,
    '00000000-0000-4000-8000-000000001105',
    '00000000-0000-4000-8000-000000001205',
    '00000000-0000-4000-8000-000000000402',
    40,
    40,
    0,
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  )
ON CONFLICT (goods_receipt_id, line_no) DO UPDATE
SET item_id = EXCLUDED.item_id,
    batch_id = EXCLUDED.batch_id,
    unit_id = EXCLUDED.unit_id,
    received_qty = EXCLUDED.received_qty,
    accepted_qty = EXCLUDED.accepted_qty,
    rejected_qty = EXCLUDED.rejected_qty,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO inventory.stock_ledger (
  id,
  org_id,
  movement_no,
  movement_type,
  direction,
  item_id,
  batch_id,
  warehouse_id,
  bin_id,
  unit_id,
  movement_qty,
  base_uom_code,
  source_qty,
  source_uom_code,
  conversion_factor,
  stock_status,
  source_doc_type,
  source_doc_id,
  created_by
)
VALUES
  (
    '00000000-0000-4000-8000-000000001401',
    '00000000-0000-4000-8000-000000000001',
    'MOV-LOCAL-0001',
    'receipt',
    'in',
    '00000000-0000-4000-8000-000000001101',
    '00000000-0000-4000-8000-000000001201',
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000001001',
    '00000000-0000-4000-8000-000000000401',
    120,
    'PCS',
    120,
    'PCS',
    1.000000,
    'available',
    'goods_receipt',
    '00000000-0000-4000-8000-000000001301',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001402',
    '00000000-0000-4000-8000-000000000001',
    'MOV-LOCAL-0002',
    'receipt',
    'in',
    '00000000-0000-4000-8000-000000001102',
    '00000000-0000-4000-8000-000000001202',
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000001001',
    '00000000-0000-4000-8000-000000000401',
    80,
    'PCS',
    80,
    'PCS',
    1.000000,
    'available',
    'goods_receipt',
    '00000000-0000-4000-8000-000000001301',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001403',
    '00000000-0000-4000-8000-000000000001',
    'MOV-LOCAL-0003',
    'receipt',
    'in',
    '00000000-0000-4000-8000-000000001103',
    '00000000-0000-4000-8000-000000001203',
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000001001',
    '00000000-0000-4000-8000-000000000401',
    60,
    'PCS',
    60,
    'PCS',
    1.000000,
    'available',
    'goods_receipt',
    '00000000-0000-4000-8000-000000001301',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001404',
    '00000000-0000-4000-8000-000000000001',
    'MOV-LOCAL-0004',
    'receipt',
    'in',
    '00000000-0000-4000-8000-000000001104',
    '00000000-0000-4000-8000-000000001204',
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000001001',
    '00000000-0000-4000-8000-000000000401',
    90,
    'PCS',
    90,
    'PCS',
    1.000000,
    'available',
    'goods_receipt',
    '00000000-0000-4000-8000-000000001301',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001405',
    '00000000-0000-4000-8000-000000000001',
    'MOV-LOCAL-0005',
    'receipt',
    'in',
    '00000000-0000-4000-8000-000000001105',
    '00000000-0000-4000-8000-000000001205',
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000001001',
    '00000000-0000-4000-8000-000000000402',
    40,
    'BOX',
    40,
    'BOX',
    1.000000,
    'available',
    'goods_receipt',
    '00000000-0000-4000-8000-000000001301',
    '00000000-0000-4000-8000-000000000101'
  )
ON CONFLICT (org_id, movement_no) DO NOTHING;

SET LOCAL erp.allow_stock_balance_write = 'on';

INSERT INTO inventory.stock_balances (
  id,
  org_id,
  item_id,
  batch_id,
  warehouse_id,
  bin_id,
  stock_status,
  base_uom_code,
  qty_on_hand,
  qty_reserved,
  qty_available,
  updated_by
)
VALUES
  (
    '00000000-0000-4000-8000-000000001501',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001101',
    '00000000-0000-4000-8000-000000001201',
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000001001',
    'available',
    'PCS',
    120,
    8,
    112,
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001502',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001102',
    '00000000-0000-4000-8000-000000001202',
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000001001',
    'available',
    'PCS',
    80,
    5,
    75,
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001503',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001103',
    '00000000-0000-4000-8000-000000001203',
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000001001',
    'available',
    'PCS',
    60,
    0,
    60,
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001504',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001104',
    '00000000-0000-4000-8000-000000001204',
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000001001',
    'available',
    'PCS',
    90,
    12,
    78,
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001505',
    '00000000-0000-4000-8000-000000000001',
    '00000000-0000-4000-8000-000000001105',
    '00000000-0000-4000-8000-000000001205',
    '00000000-0000-4000-8000-000000000801',
    '00000000-0000-4000-8000-000000001001',
    'available',
    'BOX',
    40,
    0,
    40,
    '00000000-0000-4000-8000-000000000101'
  )
ON CONFLICT (org_id, item_id, batch_id, warehouse_id, bin_id, stock_status) DO UPDATE
SET base_uom_code = EXCLUDED.base_uom_code,
    qty_on_hand = EXCLUDED.qty_on_hand,
    qty_reserved = EXCLUDED.qty_reserved,
    qty_available = EXCLUDED.qty_available,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO inventory.warehouse_daily_closings (
  id,
  org_id,
  closing_no,
  warehouse_id,
  business_date,
  shift_code,
  status,
  orders_processed_count,
  pending_task_count,
  variance_count,
  exception_note,
  created_by,
  updated_by
)
VALUES
  (
    '00000000-0000-4000-8000-000000001601',
    '00000000-0000-4000-8000-000000000001',
    'WDC-LOCAL-0001',
    '00000000-0000-4000-8000-000000000801',
    '2026-04-26',
    'day',
    'open',
    18,
    3,
    0,
    'Local demo closing board',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  ),
  (
    '00000000-0000-4000-8000-000000001602',
    '00000000-0000-4000-8000-000000000001',
    'WDC-LOCAL-0002',
    '00000000-0000-4000-8000-000000000802',
    '2026-04-26',
    'day',
    'in_review',
    6,
    1,
    1,
    'Return review pending',
    '00000000-0000-4000-8000-000000000101',
    '00000000-0000-4000-8000-000000000101'
  )
ON CONFLICT (org_id, closing_no) DO UPDATE
SET warehouse_id = EXCLUDED.warehouse_id,
    business_date = EXCLUDED.business_date,
    shift_code = EXCLUDED.shift_code,
    status = EXCLUDED.status,
    orders_processed_count = EXCLUDED.orders_processed_count,
    pending_task_count = EXCLUDED.pending_task_count,
    variance_count = EXCLUDED.variance_count,
    exception_note = EXCLUDED.exception_note,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

COMMIT;
