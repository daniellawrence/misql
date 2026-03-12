package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"gopkg.in/yaml.v3"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type tableData struct {
	Columns []string         `yaml:"columns"`
	Rows    []map[string]any `yaml:"rows"`
}

func loadDatabase(provider *memory.DbProvider, dbName string, dbPath string) (*memory.Database, error) {
	db := memory.NewDatabase(dbName)
	db.EnablePrimaryKeyIndexes()

	entries, err := os.ReadDir(dbPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		tableName := entry.Name()[:len(entry.Name())-5]
		tablePath := filepath.Join(dbPath, entry.Name())

		if err := loadTable(db, provider, tableName, tablePath); err != nil {
			log.Printf("warn: skipping table %s/%s: %v", dbName, tableName, err)
		}
	}

	return db, nil
}

func loadTable(db *memory.Database, provider *memory.DbProvider, tableName string, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var td tableData
	if err := yaml.Unmarshal(data, &td); err != nil {
		return err
	}

	schema := make(sql.Schema, len(td.Columns))
	for i, col := range td.Columns {
		schema[i] = &sql.Column{
			Name:     col,
			Type:     types.LongText,
			Nullable: true,
			Source:   tableName,
		}
	}

	table := memory.NewTable(db, tableName, sql.NewPrimaryKeySchema(schema), nil)
	db.AddTable(tableName, table)

	sess := memory.NewSession(sql.NewBaseSession(), provider)
	ctx := sql.NewContext(nil, sql.WithSession(sess))

	for _, row := range td.Rows {
		sqlRow := make(sql.Row, len(td.Columns))
		for i, col := range td.Columns {
			if v, ok := row[col]; ok {
				switch val := v.(type) {
				case []interface{}:
					b, _ := json.Marshal(val)
					sqlRow[i] = string(b)
				default:
					sqlRow[i] = fmt.Sprintf("%v", val)
				}
			}
		}
		if err := table.Insert(ctx, sqlRow); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	dataDir := envOr("MYSQL_DATA_DIR", "./data")
	listenHost := envOr("MYSQL_HOST", "0.0.0.0")
	listenPort := envOr("MYSQL_TCP_PORT", "3306")

	// tmpProvider is used only to create sessions during data loading.
	tmpProvider := memory.NewDBProvider()

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		log.Fatalf("cannot read data dir %q: %v", dataDir, err)
	}

	var databases []sql.Database
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dbPath := filepath.Join(dataDir, entry.Name())
		db, err := loadDatabase(tmpProvider, entry.Name(), dbPath)
		if err != nil {
			log.Printf("warn: skipping database %s: %v", entry.Name(), err)
			continue
		}
		log.Printf("loaded database: %s", entry.Name())
		databases = append(databases, db)
	}

	provider := memory.NewDBProvider(databases...)
	engine := sqle.NewDefault(provider)

	addr := listenHost + ":" + listenPort
	config := server.Config{
		Protocol: "tcp",
		Address:  addr,
	}

	sessionBuilder := memory.NewSessionBuilder(provider)

	s, err := server.NewServer(config, engine, sql.NewContext, sessionBuilder, nil)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	log.Printf("MiSQL listening on %s", addr)
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
