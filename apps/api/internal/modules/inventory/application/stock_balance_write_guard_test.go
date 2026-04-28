package application

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestStockBalanceWriteGuardPreventsDirectModuleMutation(t *testing.T) {
	apiRoot := apiRootPath(t)
	migrationBytes, err := os.ReadFile(filepath.Join(apiRoot, "migrations", "000002_create_phase1_base_tables.up.sql"))
	if err != nil {
		t.Fatalf("read stock balance guard migration: %v", err)
	}
	migration := string(migrationBytes)
	if !strings.Contains(migration, "current_setting('erp.allow_stock_balance_write', true)") ||
		!strings.Contains(migration, "CREATE TRIGGER trg_stock_balances_write_guard") {
		t.Fatal("stock balance migration must keep the write-context trigger guard")
	}

	directWrite := regexp.MustCompile(`(?is)\b(insert\s+into|update|delete\s+from)\s+inventory\.stock_balances\b`)
	allowedWriter := filepath.Clean(filepath.Join(apiRoot, "internal", "modules", "inventory", "application", "postgres_stock_movement_store.go"))
	scanned := 0

	err = filepath.WalkDir(filepath.Join(apiRoot, "internal", "modules"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		scanned++
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if directWrite.Match(content) && filepath.Clean(path) != allowedWriter {
			t.Fatalf("direct stock balance mutation is only allowed in %s; found in %s", allowedWriter, path)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("scan module files: %v", err)
	}
	if scanned == 0 {
		t.Fatal("stock balance guard scanned no module files")
	}
}

func apiRootPath(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
}
