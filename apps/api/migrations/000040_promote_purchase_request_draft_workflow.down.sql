BEGIN;

DROP INDEX IF EXISTS subcontract.ix_purchase_request_drafts_org_status;

UPDATE subcontract.purchase_request_drafts
SET status = 'draft'
WHERE status IN ('submitted', 'approved', 'converted_to_po', 'cancelled', 'rejected');

ALTER TABLE subcontract.purchase_request_drafts
  DROP CONSTRAINT IF EXISTS ck_purchase_request_drafts_status;

ALTER TABLE subcontract.purchase_request_drafts
  ADD CONSTRAINT ck_purchase_request_drafts_status CHECK (
    status IN ('draft')
  );

COMMIT;
