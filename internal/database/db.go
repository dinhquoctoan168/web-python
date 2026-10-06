// Package database
// Mục đích: Quản lý kết nối SQLite singleton.
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	dbInstance *sql.DB
	once       sync.Once
)

// GetDB trả về singleton kết nối SQLite
func GetDB() *sql.DB {
	return dbInstance
}

// InitDB mở và khởi tạo cơ sở dữ liệu từ đường dẫn cấu hình
func InitDB(dbPath string) (*sql.DB, error) {
	var err error
	once.Do(func() {
		if dbPath == "" {
			dbPath = "data/Database/algo_db.db"
		}

		dir := filepath.Dir(dbPath)
		if err = os.MkdirAll(dir, 0755); err != nil {
			err = fmt.Errorf("không thể tạo thư mục database: %w", err)
			return
		}

		dbInstance, err = sql.Open("sqlite", dbPath)
		if err != nil {
			err = fmt.Errorf("không thể mở sqlite: %w", err)
			return
		}

		if err = initSchema(dbInstance); err != nil {
			err = fmt.Errorf("khởi tạo schema thất bại: %w", err)
			return
		}

		if err = seedInitialData(dbInstance); err != nil {
			log.Printf("Cảnh báo seed dữ liệu: %v", err)
		}

		if err = seedFunctions(dbInstance); err != nil {
			log.Printf("Cảnh báo seed cú pháp/hàm: %v", err)
		}
	})

	return dbInstance, err
}

// CloseDB đóng kết nối database
func CloseDB() {
	if dbInstance != nil {
		dbInstance.Close()
	}
}
