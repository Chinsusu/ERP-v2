BEGIN;

INSERT INTO mdm.uoms (
  uom_code,
  name_vi,
  name_en,
  uom_group,
  decimal_scale,
  allow_decimal,
  is_global_convertible,
  description
) VALUES (
  'PACK',
  'Gói',
  'Pack',
  'PACK',
  0,
  false,
  false,
  'Item-specific pack UOM'
)
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

COMMIT;
