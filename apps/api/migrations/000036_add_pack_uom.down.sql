BEGIN;

DELETE FROM mdm.uoms
WHERE uom_code = 'PACK';

COMMIT;
