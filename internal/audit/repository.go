package audit

import (
	"database/sql"
	"time"
)

// Repository quản lý lưu trữ và truy vấn audit logs trong CSDL
type Repository struct {
	db *sql.DB
}

// NewRepository khởi tạo repository audit log
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Record ghi một bản ghi kiểm toán mới
func (r *Repository) Record(log *AuditLog) error {
	query := `INSERT INTO audit_logs (user_id, action, object_type, object_id, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	now := time.Now()
	res, err := r.db.Exec(query, log.UserID, log.Action, log.ObjectType, log.ObjectID, log.Metadata, now)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	log.ID = int(id)
	log.CreatedAt = now
	return nil
}

// ListRecent lấy danh sách nhật ký kiểm toán mới nhất
func (r *Repository) ListRecent(limit int) ([]*AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, user_id, action, object_type, object_id, COALESCE(metadata, ''), created_at
		FROM audit_logs
		ORDER BY id DESC LIMIT ?`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		var l AuditLog
		var uid, oid sql.NullInt64
		if err := rows.Scan(&l.ID, &uid, &l.Action, &l.ObjectType, &oid, &l.Metadata, &l.CreatedAt); err != nil {
			return nil, err
		}
		if uid.Valid {
			u := int(uid.Int64)
			l.UserID = &u
		}
		if oid.Valid {
			o := int(oid.Int64)
			l.ObjectID = &o
		}
		logs = append(logs, &l)
	}
	return logs, nil
}

// ListByUser lấy lịch sử hoạt động của một người dùng
func (r *Repository) ListByUser(userID, limit int) ([]*AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, user_id, action, object_type, object_id, COALESCE(metadata, ''), created_at
		FROM audit_logs
		WHERE user_id = ?
		ORDER BY id DESC LIMIT ?`
	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		var l AuditLog
		var uid, oid sql.NullInt64
		if err := rows.Scan(&l.ID, &uid, &l.Action, &l.ObjectType, &oid, &l.Metadata, &l.CreatedAt); err != nil {
			return nil, err
		}
		if uid.Valid {
			u := int(uid.Int64)
			l.UserID = &u
		}
		if oid.Valid {
			o := int(oid.Int64)
			l.ObjectID = &o
		}
		logs = append(logs, &l)
	}
	return logs, nil
}

// ListByObject lấy nhật ký thao tác trên một đối tượng cụ thể
func (r *Repository) ListByObject(objectType string, objectID, limit int) ([]*AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, user_id, action, object_type, object_id, COALESCE(metadata, ''), created_at
		FROM audit_logs
		WHERE object_type = ? AND object_id = ?
		ORDER BY id DESC LIMIT ?`
	rows, err := r.db.Query(query, objectType, objectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		var l AuditLog
		var uid, oid sql.NullInt64
		if err := rows.Scan(&l.ID, &uid, &l.Action, &l.ObjectType, &oid, &l.Metadata, &l.CreatedAt); err != nil {
			return nil, err
		}
		if uid.Valid {
			u := int(uid.Int64)
			l.UserID = &u
		}
		if oid.Valid {
			o := int(oid.Int64)
			l.ObjectID = &o
		}
		logs = append(logs, &l)
	}
	return logs, nil
}
