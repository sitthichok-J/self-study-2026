package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/model"
)

// ทุก query เป็น parameterized ($1, $2, ...) ไม่เอาค่าจากผู้ใช้มาต่อ string เด็ดขาด (กัน SQL injection)
const (
	qInsert = `INSERT INTO claims (patient_ref, hospital, amount_satang, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	qGet = `SELECT id, patient_ref, hospital, amount_satang, status, created_at, updated_at
FROM claims WHERE id = $1`
	qList = `SELECT id, patient_ref, hospital, amount_satang, status, created_at, updated_at
FROM claims WHERE ($1::text = '' OR status = $1::text) ORDER BY id LIMIT NULLIF($2::int, 0) OFFSET $3::int`
	qUpdate = `UPDATE claims SET patient_ref = $2, hospital = $3, amount_satang = $4, status = $5, updated_at = $6
WHERE id = $1`
	qDelete    = `DELETE FROM claims WHERE id = $1`
	qSetStatus = `UPDATE claims SET status = $2, updated_at = $3 WHERE id = $1`
)

// Queries ให้สคริปต์ตรวจ SQL (scripts/check_sql.sh) ใช้ query ชุดเดียวกับโค้ดจริง
func Queries() map[string]string {
	return map[string]string{
		"insert": qInsert, "get": qGet, "list": qList,
		"update": qUpdate, "delete": qDelete, "set_status": qSetStatus,
	}
}

// Postgres ใช้ database/sql จึงใช้ได้กับ driver ใดก็ได้ (ในโปรเจกต์ใช้ pgx stdlib ผ่าน build tag pgx)
type Postgres struct{ db *sql.DB }

func NewPostgres(db *sql.DB) *Postgres { return &Postgres{db: db} }

func (p *Postgres) Create(ctx context.Context, c model.Claim) (model.Claim, error) {
	err := p.db.QueryRowContext(ctx, qInsert,
		c.PatientRef, c.Hospital, c.AmountSatang, string(c.Status), c.CreatedAt, c.UpdatedAt).Scan(&c.ID)
	if err != nil {
		return model.Claim{}, fmt.Errorf("insert claim: %w", err)
	}
	return c, nil
}

func scan(row interface{ Scan(...any) error }) (model.Claim, error) {
	var c model.Claim
	var status string
	err := row.Scan(&c.ID, &c.PatientRef, &c.Hospital, &c.AmountSatang, &status, &c.CreatedAt, &c.UpdatedAt)
	c.Status = model.Status(status)
	return c, err
}

func (p *Postgres) Get(ctx context.Context, id int64) (model.Claim, error) {
	c, err := scan(p.db.QueryRowContext(ctx, qGet, id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.Claim{}, model.ErrNotFound
	}
	if err != nil {
		return model.Claim{}, fmt.Errorf("get claim: %w", err)
	}
	return c, nil
}

func (p *Postgres) List(ctx context.Context, f model.ListFilter) ([]model.Claim, error) {
	rows, err := p.db.QueryContext(ctx, qList, string(f.Status), f.Limit, f.Offset)
	if err != nil {
		return nil, fmt.Errorf("list claims: %w", err)
	}
	defer rows.Close()
	out := []model.Claim{}
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan claim: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list claims: %w", err)
	}
	return out, nil
}

func affected(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (p *Postgres) Update(ctx context.Context, c model.Claim) error {
	err := affected(p.db.ExecContext(ctx, qUpdate, c.ID, c.PatientRef, c.Hospital, c.AmountSatang, string(c.Status), c.UpdatedAt))
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return fmt.Errorf("update claim: %w", err)
	}
	return err
}

func (p *Postgres) Delete(ctx context.Context, id int64) error {
	err := affected(p.db.ExecContext(ctx, qDelete, id))
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return fmt.Errorf("delete claim: %w", err)
	}
	return err
}

// UpdateStatusMany ทำใน transaction เดียว: ถ้ามีรายการใดไม่พบหรือพลาด rollback ทั้งหมด
func (p *Postgres) UpdateStatusMany(ctx context.Context, ids []int64, s model.Status, at time.Time) (err error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback() // rollback ผ่าน defer เมื่อมี error (Rollback หลัง Commit แล้วไม่เป็นไร)
		}
	}()
	for _, id := range ids {
		if err = affected(tx.ExecContext(ctx, qSetStatus, id, string(s), at)); err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return fmt.Errorf("claim %d: %w", id, err)
			}
			return fmt.Errorf("set status: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
