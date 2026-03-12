package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/sql"
)

func TestLoadTable_ListToJSON(t *testing.T) {
	yaml := `
columns:
  - name
  - tags
rows:
  - name: "example"
    tags:
      - foo
      - bar
      - baz
`
	dir := t.TempDir()
	path := filepath.Join(dir, "things.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	provider := memory.NewDBProvider()
	db := memory.NewDatabase("test")
	db.EnablePrimaryKeyIndexes()

	if err := loadTable(db, provider, "things", path); err != nil {
		t.Fatalf("loadTable: %v", err)
	}

	sess := memory.NewSession(sql.NewBaseSession(), provider)
	ctx := sql.NewContext(nil, sql.WithSession(sess))

	tbl, ok, err := db.GetTableInsensitive(ctx, "things")
	if err != nil || !ok {
		t.Fatalf("table not found: %v", err)
	}

	piter, err := tbl.(sql.Table).Partitions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	part, err := piter.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}

	riter, err := tbl.(sql.Table).PartitionRows(ctx, part)
	if err != nil {
		t.Fatal(err)
	}
	row, err := riter.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// column 0 = name, column 1 = tags
	got, _ := row[1].(string)
	want := `["foo","bar","baz"]`
	if got != want {
		t.Errorf("tags: got %q, want %q", got, want)
	}
}
