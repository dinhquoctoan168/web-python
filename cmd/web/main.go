// Package main
// Mục đích: Khởi chạy máy chủ HTTP thuần thư viện chuẩn Go, nạp cấu hình và định tuyến URL.
package main

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"web_python/internal/database"
	"web_python/internal/frontend"
)

// loadEnv đọc tệp cấu hình .env thủ công bằng standard library để không phụ thuộc lib ngoài
func loadEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("Thông báo: Không tìm thấy tệp env %s: %v", filepath, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}
}

func main() {
	log.Println("==================================================")
	log.Println("  Hệ thống học tập Web Python IDE (CSDL & Giải thuật) ")
	log.Println("==================================================")

	// 1. Nạp biến môi trường
	loadEnv("config/.env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/Database/algo_db.db"
	}

	// 2. Khởi tạo cơ sở dữ liệu
	_, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Khởi tạo database thất bại: %v", err)
	}
	defer database.CloseDB()

	// 3. Thiết lập Mux định tuyến thuần standard library
	mux := http.NewServeMux()

	// Phục vụ tài nguyên tĩnh (CSS, JS, Fonts)
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Trang giao diện IDE
	mux.HandleFunc("/", frontend.HandleIDE)
	mux.HandleFunc("/ide", frontend.HandleIDE)

	// API endpoints
	mux.HandleFunc("/api/exercise", frontend.HandleAPIExercise)
	mux.HandleFunc("/api/functions", frontend.HandleAPIFunctions)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Xử lý dừng máy chủ an toàn (Graceful Shutdown)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server đang chạy tại http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Lỗi máy chủ HTTP: %v", err)
		}
	}()

	<-stop
	log.Println("\nĐang tắt máy chủ an toàn...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Lỗi khi shutdown: %v", err)
	}
	log.Println("Máy chủ đã dừng hoàn tất.")
}
