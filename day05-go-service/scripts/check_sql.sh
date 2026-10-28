#!/usr/bin/env bash
# ตรวจ migration และ SQL ของ repository กับ Postgres จริง (ไม่ต้องใช้ Go driver)
# ใช้: PGURL=postgres://claims:claims@localhost:5433/claims?sslmode=disable ./scripts/check_sql.sh
# ต้องมี psql และ go; สคริปต์จะ drop/create ตาราง claims (ใช้กับ DB ทดสอบเท่านั้น)
set -euo pipefail
cd "$(dirname "$0")/.."
: "${PGURL:?ตั้ง PGURL ก่อน}"
P="psql $PGURL -v ON_ERROR_STOP=1 -q -t -A"

echo "== migrate down/up"
$P -f migrations/000001_create_claims.down.sql
$P -f migrations/000001_create_claims.up.sql

SQL=$(go run ./cmd/printsql)
q() { echo "$SQL" | awk -v n="$1" '$0=="-- "n{f=1;next} /^-- /{f=0} f'; }

prep() { # name  types
  echo "PREPARE $1($2) AS $(q "$1")"
}
echo "== queries (query ชุดเดียวกับที่โค้ด Go ใช้)"
out=$($P <<SQL
$(prep insert "text,text,bigint,text,timestamptz,timestamptz")
$(prep get "bigint")
$(prep list "text,int,int")
$(prep update "bigint,text,text,bigint,text,timestamptz")
$(prep set_status "bigint,text,timestamptz")
$(prep delete "bigint")
EXECUTE insert('P1','H1',1000,'PENDING',now(),now());
EXECUTE insert('P2','H2',2000,'APPROVED',now(),now());
EXECUTE insert('P3','H1',3000,'PENDING',now(),now());
\echo get1
EXECUTE get(1);
\echo list_all
EXECUTE list('',0,0);
\echo list_pending_limit1
EXECUTE list('PENDING',1,0);
\echo list_offset1_limit1
EXECUTE list('',1,1);
\echo update
EXECUTE update(1,'P1x','H9',5000,'APPROVED',now());
EXECUTE get(1);
\echo set_status
EXECUTE set_status(3,'REJECTED',now());
EXECUTE get(3);
\echo delete
EXECUTE delete(2);
EXECUTE get(2);
SQL
)
echo "$out"

expect() { # ชื่อ  ค่าที่ต้องปรากฏ
  if ! grep -q -- "$2" <<<"$out"; then echo "FAIL: $1 (ไม่พบ '$2')"; exit 1; fi
  echo "ok: $1"
}
expect "insert/get id 1"        "^1|P1|H1|1000|PENDING"
expect "list ทั้งหมด 3 แถว"      "^3|P3|H1|3000|PENDING"
expect "update"                 "^1|P1x|H9|5000|APPROVED"
expect "set_status"             "^3|P3|H1|3000|REJECTED"
if grep -q "^2|P2" <<<"$(echo "$out" | sed -n '/^delete$/,$p')"; then echo "FAIL: delete"; exit 1; fi
echo "ok: delete"

echo "== constraints ต้องกันข้อมูลผิด"
for bad in \
  "INSERT INTO claims (patient_ref,hospital,amount_satang,status,created_at,updated_at) VALUES ('P','H',0,'PENDING',now(),now())" \
  "INSERT INTO claims (patient_ref,hospital,amount_satang,status,created_at,updated_at) VALUES ('P','H',10,'WRONG',now(),now())" \
  "INSERT INTO claims (patient_ref,hospital,amount_satang,status,created_at,updated_at) VALUES (NULL,'H',10,'PENDING',now(),now())"; do
  if $P -c "$bad" >/dev/null 2>&1; then echo "FAIL: ควรถูกปฏิเสธ: $bad"; exit 1; fi
  echo "ok: ปฏิเสธ"
done

echo "== tx rollback: ส่วนที่แก้ไปแล้วต้องหายเมื่อ rollback"
$P -c "TRUNCATE claims RESTART IDENTITY" >/dev/null
$P -c "INSERT INTO claims (patient_ref,hospital,amount_satang,status,created_at,updated_at) VALUES ('A','H',1,'PENDING',now(),now())" >/dev/null
$P <<'SQL' >/dev/null
BEGIN;
UPDATE claims SET status='APPROVED' WHERE id=1;
ROLLBACK;
SQL
st=$($P -c "SELECT status FROM claims WHERE id=1")
[ "$st" = "PENDING" ] && echo "ok: rollback" || { echo "FAIL: rollback ($st)"; exit 1; }

echo "== migrate down"
$P -f migrations/000001_create_claims.down.sql
echo "ALL OK"
