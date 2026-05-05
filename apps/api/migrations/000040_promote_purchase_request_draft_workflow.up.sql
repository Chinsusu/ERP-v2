BEGIN;

ALTER TABLE subcontract.purchase_request_drafts
  DROP CONSTRAINT IF EXISTS ck_purchase_request_drafts_status;

ALTER TABLE subcontract.purchase_request_drafts
  ADD CONSTRAINT ck_purchase_request_drafts_status CHECK (
    status IN ('draft', 'submitted', 'approved', 'converted_to_po', 'cancelled', 'rejected')
  );

CREATE INDEX IF NOT EXISTS ix_purchase_request_drafts_org_status
  ON subcontract.purchase_request_drafts(org_id, status, created_at DESC);

COMMIT;
