BEGIN;

CREATE TABLE mdm.uoms (
  uom_code varchar(20) PRIMARY KEY,
  name_vi text NOT NULL,
  name_en text NOT NULL,
  uom_group varchar(30) NOT NULL,
  decimal_scale integer NOT NULL DEFAULT 0,
  allow_decimal boolean NOT NULL DEFAULT false,
  is_global_convertible boolean NOT NULL DEFAULT false,
  is_active boolean NOT NULL DEFAULT true,
  description text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_uoms_code CHECK (uom_code ~ '^[A-Z0-9_-]+$'),
  CONSTRAINT ck_uoms_group CHECK (uom_group IN ('MASS', 'VOLUME', 'COUNT', 'PACK', 'SERVICE')),
  CONSTRAINT ck_uoms_decimal_scale CHECK (decimal_scale BETWEEN 0 AND 6),
  CONSTRAINT ck_uoms_decimal_allowed CHECK (allow_decimal OR decimal_scale = 0)
);

CREATE TABLE mdm.uom_conversions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  item_id uuid REFERENCES mdm.items(id),
  from_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  to_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  factor numeric(18, 6) NOT NULL,
  conversion_type varchar(30) NOT NULL,
  effective_from date NOT NULL DEFAULT CURRENT_DATE,
  effective_to date,
  is_active boolean NOT NULL DEFAULT true,
  created_by uuid REFERENCES core.users(id),
  approved_by uuid REFERENCES core.users(id),
  approved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_uom_conversions_factor_positive CHECK (factor > 0),
  CONSTRAINT ck_uom_conversions_type CHECK (conversion_type IN ('GLOBAL', 'ITEM_SPECIFIC')),
  CONSTRAINT uq_uom_conversions_scope UNIQUE NULLS NOT DISTINCT (item_id, from_uom_code, to_uom_code),
  CONSTRAINT ck_uom_conversions_scope CHECK (
    (conversion_type = 'GLOBAL' AND item_id IS NULL)
    OR (conversion_type = 'ITEM_SPECIFIC' AND item_id IS NOT NULL)
  ),
  CONSTRAINT ck_uom_conversions_effective_dates CHECK (effective_to IS NULL OR effective_to >= effective_from)
);

CREATE INDEX ix_uom_conversions_active_lookup
  ON mdm.uom_conversions(from_uom_code, to_uom_code, is_active);

INSERT INTO mdm.uoms (
  uom_code,
  name_vi,
  name_en,
  uom_group,
  decimal_scale,
  allow_decimal,
  is_global_convertible,
  description
) VALUES
  ('MG', 'Milligram', 'Milligram', 'MASS', 6, true, true, 'R&D and small BOM quantity'),
  ('G', 'Gram', 'Gram', 'MASS', 6, true, true, 'Base UOM for solid raw materials'),
  ('KG', 'Kilogram', 'Kilogram', 'MASS', 6, true, true, 'Purchase UOM for solid raw materials'),
  ('ML', 'Milliliter', 'Milliliter', 'VOLUME', 6, true, true, 'Base UOM for liquid materials'),
  ('L', 'Liter', 'Liter', 'VOLUME', 6, true, true, 'Purchase UOM for liquid materials'),
  ('PCS', 'Piece', 'Piece', 'COUNT', 0, false, false, 'Count unit'),
  ('BOTTLE', 'Chai', 'Bottle', 'PACK', 0, false, false, 'Item-specific package UOM'),
  ('JAR', 'Hũ', 'Jar', 'PACK', 0, false, false, 'Item-specific package UOM'),
  ('TUBE', 'Tuýp', 'Tube', 'PACK', 0, false, false, 'Item-specific package UOM'),
  ('BOX', 'Hộp', 'Box', 'PACK', 0, false, false, 'Item-specific package UOM'),
  ('CARTON', 'Thùng', 'Carton', 'PACK', 0, false, false, 'Item-specific package UOM'),
  ('SET', 'Bộ/combo', 'Set', 'PACK', 0, false, false, 'BOM/combo UOM'),
  ('SERVICE', 'Dịch vụ', 'Service', 'SERVICE', 0, false, false, 'Non-stock service')
ON CONFLICT (uom_code) DO UPDATE SET
  name_vi = EXCLUDED.name_vi,
  name_en = EXCLUDED.name_en,
  uom_group = EXCLUDED.uom_group,
  decimal_scale = EXCLUDED.decimal_scale,
  allow_decimal = EXCLUDED.allow_decimal,
  is_global_convertible = EXCLUDED.is_global_convertible,
  description = EXCLUDED.description,
  is_active = true,
  updated_at = now();

INSERT INTO mdm.uom_conversions (
  from_uom_code,
  to_uom_code,
  factor,
  conversion_type
) VALUES
  ('KG', 'G', 1000.000000, 'GLOBAL'),
  ('G', 'KG', 0.001000, 'GLOBAL'),
  ('MG', 'G', 0.001000, 'GLOBAL'),
  ('G', 'MG', 1000.000000, 'GLOBAL'),
  ('L', 'ML', 1000.000000, 'GLOBAL'),
  ('ML', 'L', 0.001000, 'GLOBAL')
ON CONFLICT ON CONSTRAINT uq_uom_conversions_scope DO UPDATE SET
  factor = EXCLUDED.factor,
  conversion_type = EXCLUDED.conversion_type,
  is_active = true,
  updated_at = now();

COMMIT;
