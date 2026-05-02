BEGIN;

DROP TABLE IF EXISTS finance.cash_transaction_allocations;
DROP TABLE IF EXISTS finance.cash_transactions;
DROP TABLE IF EXISTS finance.cod_discrepancies;
DROP TABLE IF EXISTS finance.cod_remittance_lines;
DROP TABLE IF EXISTS finance.cod_remittances;
DROP TABLE IF EXISTS finance.supplier_payable_lines;
DROP TABLE IF EXISTS finance.supplier_payables;
DROP TABLE IF EXISTS finance.customer_receivable_lines;
DROP TABLE IF EXISTS finance.customer_receivables;
DROP SCHEMA IF EXISTS finance;

COMMIT;
