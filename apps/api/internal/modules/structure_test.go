package modules

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var phase1Modules = []string{
	"masterdata",
	"purchase",
	"inventory",
	"qc",
	"production",
	"sales",
	"shipping",
	"returns",
	"finance",
	"reporting",
}

var moduleComponents = []string{
	"handler",
	"application",
	"domain",
	"repository",
	"dto",
	"events",
	"queries",
	"tests",
}

func TestPhase1ModuleFoldersExist(t *testing.T) {
	root := moduleRoot(t)

	for _, module := range phase1Modules {
		for _, component := range moduleComponents {
			path := filepath.Join(root, module, component)
			if !isDirectory(path) {
				t.Fatalf("%s/%s folder is missing", module, component)
			}
		}
	}
}

func TestModulesDoNotImportOtherModuleRepositories(t *testing.T) {
	root := moduleRoot(t)
	const moduleImportPrefix = "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/"

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		owner, ok := ownerModule(root, path)
		if !ok {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, spec := range file.Imports {
			importPath := strings.Trim(spec.Path.Value, `"`)
			if !strings.HasPrefix(importPath, moduleImportPrefix) {
				continue
			}

			relativeImport := strings.TrimPrefix(importPath, moduleImportPrefix)
			parts := strings.Split(relativeImport, "/")
			if len(parts) < 2 || parts[1] != "repository" {
				continue
			}

			importedModule := parts[0]
			if importedModule != owner {
				t.Fatalf("%s imports %s repository directly via %s", owner, importedModule, importPath)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("scan module imports: %v", err)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve module structure test path")
	}

	return filepath.Dir(file)
}

func ownerModule(root string, path string) (string, bool) {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return "", false
	}

	parts := strings.Split(relative, string(filepath.Separator))
	if len(parts) < 2 {
		return "", false
	}

	return parts[0], true
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
