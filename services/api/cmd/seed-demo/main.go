// Development-only demo data seed.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

const (
	orgID        = "22222222-2222-4222-8222-222222222222"
	ketuaID      = "22222222-2222-4222-8222-222222222301"
	sekretarisID = "22222222-2222-4222-8222-222222222302"
	bendaharaID  = "22222222-2222-4222-8222-222222222303"
	pengurusID   = "22222222-2222-4222-8222-222222222304"
	wargaBudiID  = "22222222-2222-4222-8222-222222222305"
	wargaSitiID  = "22222222-2222-4222-8222-222222222306"

	rumahA01ID = "22222222-2222-4222-8222-222222222401"
	rumahA02ID = "22222222-2222-4222-8222-222222222402"
	rumahB01ID = "22222222-2222-4222-8222-222222222403"

	budiID = "22222222-2222-4222-8222-222222222501"
	sitiID = "22222222-2222-4222-8222-222222222502"
	andiID = "22222222-2222-4222-8222-222222222503"
	rinaID = "22222222-2222-4222-8222-222222222504"
	dewiID = "22222222-2222-4222-8222-222222222505"

	keluargaBudiID = "22222222-2222-4222-8222-222222222601"
	keluargaRinaID = "22222222-2222-4222-8222-222222222602"
	keluargaDewiID = "22222222-2222-4222-8222-222222222603"

	iuranKeamananID   = "22222222-2222-4222-8222-222222222701"
	iuranKebersihanID = "22222222-2222-4222-8222-222222222702"
	iuranSosialID     = "22222222-2222-4222-8222-222222222703"

	tagihanLunasID   = "22222222-2222-4222-8222-222222222801"
	tagihanTertundaID = "22222222-2222-4222-8222-222222222802"
	tagihanSebagianID = "22222222-2222-4222-8222-222222222803"
	pembayaranLunasID = "22222222-2222-4222-8222-222222222901"

	kategoriMasukID      = "22222222-2222-4222-8222-222222223001"
	kategoriOperasionalID = "22222222-2222-4222-8222-222222223002"
	kategoriSosialID      = "22222222-2222-4222-8222-222222223003"

	pengumumanID = "22222222-2222-4222-8222-222222223101"
	agendaID     = "22222222-2222-4222-8222-222222223102"

	suratDomisiliID = "22222222-2222-4222-8222-222222223201"
	suratUsahaID    = "22222222-2222-4222-8222-222222223202"
	permohonanID    = "22222222-2222-4222-8222-222222223203"

	keluhanKategoriID = "22222222-2222-4222-8222-222222223301"
	keluhanID         = "22222222-2222-4222-8222-222222223302"
)

type demoUser struct {
	id, residentID, email, role string
}

func main() {
	if os.Getenv("APP_ENV") != "development" {
		log.Fatal("demo seeder only runs when APP_ENV=development")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer conn.Close(ctx)

	passwordHash, err := auth.HashPassword("Demo12345!")
	if err != nil {
		log.Fatalf("hash demo password: %v", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("begin transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
INSERT INTO organizations (id, name, rt_number, rw_number, address, timezone, status)
VALUES ($1, 'RT 02 Taman Melati', '02', '03',
        'Jl. Melati Raya, Kelurahan Sukamaju, Kecamatan Cimanggis, Kota Depok', 'Asia/Jakarta', 'active')
ON CONFLICT (id) DO NOTHING`, orgID); err != nil {
		log.Fatalf("seed organization: %v", err)
	}
	if err := seedUsers(ctx, tx, passwordHash, false); err != nil {
		log.Fatalf("seed demo users: %v", err)
	}
	if err := seedDomain(ctx, tx); err != nil {
		log.Fatalf("seed domain data: %v", err)
	}
	if err := seedUsers(ctx, tx, passwordHash, true); err != nil {
		log.Fatalf("link demo users to residents: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit seed: %v", err)
	}

	fmt.Println("Demo seeded: organisasi RT 02/RW 03 Taman Melati; semua akun memakai password Demo12345!")
}

func seedDomain(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
INSERT INTO organizations (id, name, rt_number, rw_number, address, timezone, status)
VALUES ('22222222-2222-4222-8222-222222222222', 'RT 02 Taman Melati', '02', '03',
        'Jl. Melati Raya, Kelurahan Sukamaju, Kecamatan Cimanggis, Kota Depok', 'Asia/Jakarta', 'active')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name, rt_number = EXCLUDED.rt_number, rw_number = EXCLUDED.rw_number,
    address = EXCLUDED.address, timezone = EXCLUDED.timezone, status = EXCLUDED.status;

INSERT INTO house_units (id, organization_id, code, address_detail, occupancy_status, status) VALUES
    ('22222222-2222-4222-8222-222222222401', '22222222-2222-4222-8222-222222222222', 'A-01', 'Blok A Nomor 01', 'owned', 'active'),
    ('22222222-2222-4222-8222-222222222402', '22222222-2222-4222-8222-222222222222', 'A-02', 'Blok A Nomor 02', 'rented', 'active'),
    ('22222222-2222-4222-8222-222222222403', '22222222-2222-4222-8222-222222222222', 'B-01', 'Blok B Nomor 01', 'contract', 'active')
ON CONFLICT (organization_id, code) DO NOTHING;

INSERT INTO residents (
    id, organization_id, full_name, birth_place, birth_date, gender, occupation,
    phone, email, resident_status, verification_status, education_level_id, marital_status_id
)
SELECT v.id, '22222222-2222-4222-8222-222222222222', v.full_name, v.birth_place, v.birth_date, v.gender,
       v.occupation, v.phone, v.email, 'active', 'verified', el.id, ms.id
FROM (VALUES
    ('22222222-2222-4222-8222-222222222501'::uuid, 'Budi Santoso', 'Jakarta', DATE '1984-08-17', 'male', 'Karyawan swasta', '081234567801', 'budi.demo@example.test', 's1', 'married'),
    ('22222222-2222-4222-8222-222222222502'::uuid, 'Siti Aminah', 'Bandung', DATE '1987-04-21', 'female', 'Guru', '081234567802', 'siti.demo@example.test', 's1', 'married'),
    ('22222222-2222-4222-8222-222222222503'::uuid, 'Andi Pratama', 'Depok', DATE '2012-11-03', 'male', 'Pelajar', NULL, NULL, 'sd', 'single'),
    ('22222222-2222-4222-8222-222222222504'::uuid, 'Rina Wulandari', 'Bogor', DATE '1991-01-15', 'female', 'Wiraswasta', '081234567804', 'rina.demo@example.test', 'sma', 'married'),
    ('22222222-2222-4222-8222-222222222505'::uuid, 'Dewi Lestari', 'Bekasi', DATE '1994-06-08', 'female', 'Desainer grafis', '081234567805', 'dewi.demo@example.test', 'diploma', 'single')
) AS v(id, full_name, birth_place, birth_date, gender, occupation, phone, email, education_code, marital_code)
JOIN education_levels el ON el.code = v.education_code
JOIN marital_statuses ms ON ms.code = v.marital_code
ON CONFLICT (organization_id, id) DO NOTHING;

INSERT INTO households (
    id, organization_id, house_unit_id, internal_number, domicile_status, move_in_date,
    verification_status, domicile_review_due_at, domicile_last_confirmed_at
) VALUES
    ('22222222-2222-4222-8222-222222222601', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222401', 'KK-RT02-001', 'permanent', DATE '2018-01-01', 'verified', CURRENT_DATE + 30, now() - INTERVAL '11 months'),
    ('22222222-2222-4222-8222-222222222602', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222402', 'KK-RT02-002', 'temporary', DATE '2025-07-01', 'verified', CURRENT_DATE + 7, now() - INTERVAL '5 months'),
    ('22222222-2222-4222-8222-222222222603', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222403', 'KK-RT02-003', 'permanent', DATE '2023-02-01', 'unverified', CURRENT_DATE - 3, NULL)
ON CONFLICT (organization_id, internal_number) DO NOTHING;

INSERT INTO household_members (id, organization_id, household_id, resident_id, relationship, is_active, started_at) VALUES
    ('22222222-2222-4222-8222-222222222611', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222601', '22222222-2222-4222-8222-222222222501', 'head', true, DATE '2018-01-01'),
    ('22222222-2222-4222-8222-222222222612', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222601', '22222222-2222-4222-8222-222222222502', 'spouse', true, DATE '2018-01-01'),
    ('22222222-2222-4222-8222-222222222613', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222601', '22222222-2222-4222-8222-222222222503', 'child', true, DATE '2018-01-01'),
    ('22222222-2222-4222-8222-222222222614', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222602', '22222222-2222-4222-8222-222222222504', 'head', true, DATE '2025-07-01'),
    ('22222222-2222-4222-8222-222222222615', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222603', '22222222-2222-4222-8222-222222222505', 'head', true, DATE '2023-02-01')
ON CONFLICT (organization_id, id) DO NOTHING;

UPDATE households
SET head_resident_id = CASE id
    WHEN '22222222-2222-4222-8222-222222222601'::uuid THEN '22222222-2222-4222-8222-222222222501'::uuid
    WHEN '22222222-2222-4222-8222-222222222602'::uuid THEN '22222222-2222-4222-8222-222222222504'::uuid
    WHEN '22222222-2222-4222-8222-222222222603'::uuid THEN '22222222-2222-4222-8222-222222222505'::uuid
END
WHERE organization_id = '22222222-2222-4222-8222-222222222222'
  AND id IN (
    '22222222-2222-4222-8222-222222222601',
    '22222222-2222-4222-8222-222222222602',
    '22222222-2222-4222-8222-222222222603'
  );

INSERT INTO due_types (
    id, organization_id, name, description, amount, frequency, due_day, status,
    automatic_generation_enabled, generation_lead_days, reminder_lead_days
) VALUES
    ('22222222-2222-4222-8222-222222222701', '22222222-2222-4222-8222-222222222222', 'Iuran Keamanan', 'Operasional keamanan lingkungan dan ronda malam.', 25000, 'monthly', 10, 'active', true, 7, 3),
    ('22222222-2222-4222-8222-222222222702', '22222222-2222-4222-8222-222222222222', 'Iuran Kebersihan', 'Pengangkutan sampah dan kebersihan lingkungan.', 20000, 'monthly', 10, 'active', true, 7, 3),
    ('22222222-2222-4222-8222-222222222703', '22222222-2222-4222-8222-222222222222', 'Dana Sosial', 'Dana kegiatan sosial warga.', 50000, 'once', NULL, 'active', false, 0, 3)
ON CONFLICT (organization_id, name) DO NOTHING;

INSERT INTO invoices (
    id, organization_id, household_id, due_type_id, invoice_number, period_start, period_end, due_date,
    amount, paid_amount, status, bulk_generation_key
) VALUES
    ('22222222-2222-4222-8222-222222222801', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222601', '22222222-2222-4222-8222-222222222701', 'INV-RT02-2026-0001', date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date, (date_trunc('month', CURRENT_DATE)::date - 1), CURRENT_DATE - 20, 25000, 25000, 'paid', 'demo-previous-month'),
    ('22222222-2222-4222-8222-222222222802', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222602', '22222222-2222-4222-8222-222222222702', 'INV-RT02-2026-0002', date_trunc('month', CURRENT_DATE)::date, (date_trunc('month', CURRENT_DATE + INTERVAL '1 month')::date - 1), CURRENT_DATE + 3, 20000, 0, 'unpaid', 'demo-current-month'),
    ('22222222-2222-4222-8222-222222222803', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222603', '22222222-2222-4222-8222-222222222701', 'INV-RT02-2026-0003', date_trunc('month', CURRENT_DATE)::date, (date_trunc('month', CURRENT_DATE + INTERVAL '1 month')::date - 1), CURRENT_DATE + 3, 25000, 10000, 'partial', 'demo-current-month')
ON CONFLICT (organization_id, invoice_number) DO NOTHING;

INSERT INTO payments (
    id, organization_id, invoice_id, payment_number, method, amount, paid_at, verification_status,
    verified_by, verified_at, created_by, idempotency_key
) VALUES
    ('22222222-2222-4222-8222-222222222901', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222801', 'PAY-RT02-2026-0001', 'cash', 25000, now() - INTERVAL '12 days', 'verified',
     '22222222-2222-4222-8222-222222222303', now() - INTERVAL '11 days', '22222222-2222-4222-8222-222222222305', 'demo-payment-001')
ON CONFLICT (organization_id, payment_number) DO NOTHING;

INSERT INTO payment_allocations (id, organization_id, payment_id, invoice_id, amount)
VALUES ('22222222-2222-4222-8222-222222222911', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222901', '22222222-2222-4222-8222-222222222801', 25000)
ON CONFLICT (organization_id, payment_id, invoice_id) DO NOTHING;

INSERT INTO cash_categories (id, organization_id, name, type, status) VALUES
    ('22222222-2222-4222-8222-222222223001', '22222222-2222-4222-8222-222222222222', 'Penerimaan Iuran', 'income', 'active'),
    ('22222222-2222-4222-8222-222222223002', '22222222-2222-4222-8222-222222222222', 'Operasional Lingkungan', 'expense', 'active'),
    ('22222222-2222-4222-8222-222222223003', '22222222-2222-4222-8222-222222222222', 'Kegiatan Sosial', 'expense', 'active')
ON CONFLICT (organization_id, type, name) DO NOTHING;

INSERT INTO cash_transactions (
    id, organization_id, transaction_number, type, category_id, amount, transaction_date, description,
    reference_type, reference_id, status, created_by
) VALUES
    ('22222222-2222-4222-8222-222222223011', '22222222-2222-4222-8222-222222222222', 'KAS-RT02-2026-0001', 'income', '22222222-2222-4222-8222-222222223001', 25000, CURRENT_DATE - 11, 'Penerimaan Iuran Keamanan dari KK-RT02-001.', 'payment', '22222222-2222-4222-8222-222222222901', 'active', '22222222-2222-4222-8222-222222222303'),
    ('22222222-2222-4222-8222-222222223012', '22222222-2222-4222-8222-222222222222', 'KAS-RT02-2026-0002', 'expense', '22222222-2222-4222-8222-222222223002', 150000, CURRENT_DATE - 5, 'Pembelian perlengkapan kebersihan lingkungan.', NULL, NULL, 'active', '22222222-2222-4222-8222-222222222303')
ON CONFLICT (organization_id, transaction_number) DO NOTHING;

INSERT INTO announcements (
    id, organization_id, author_user_id, title, content, category, priority, publish_at, expire_at, status
) VALUES (
    '22222222-2222-4222-8222-222222223101', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222301',
    'Kerja Bakti Bulanan', 'Warga diharapkan mengikuti kerja bakti pada Minggu pagi. Bawa alat kebersihan pribadi bila tersedia.',
    'event', 'important', now() - INTERVAL '1 day', now() + INTERVAL '14 days', 'published'
) ON CONFLICT (organization_id, id) DO NOTHING;

INSERT INTO announcement_targets (id, organization_id, announcement_id, target_type, target_id)
VALUES ('22222222-2222-4222-8222-222222223111', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222223101', 'all', NULL)
ON CONFLICT (organization_id, announcement_id, target_type, target_id) DO NOTHING;

INSERT INTO events (id, organization_id, author_user_id, title, description, location, starts_at, ends_at, status)
VALUES (
    '22222222-2222-4222-8222-222222223102', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222301',
    'Kerja Bakti Lingkungan', 'Pembersihan saluran air dan area taman RT.', 'Titik kumpul Pos RT 02',
    date_trunc('day', now() + INTERVAL '5 days') + INTERVAL '07:00', date_trunc('day', now() + INTERVAL '5 days') + INTERVAL '10:00', 'planned'
) ON CONFLICT (organization_id, id) DO NOTHING;

INSERT INTO letter_types (id, organization_id, name, requirements, form_schema, template, number_pattern, status, sla_hours) VALUES
    ('22222222-2222-4222-8222-222222223201', '22222222-2222-4222-8222-222222222222', 'Surat Keterangan Domisili',
     '["KTP","Kartu Keluarga"]'::jsonb, '{"keperluan":{"type":"string","label":"Keperluan"}}'::jsonb,
     'Surat keterangan domisili untuk {{full_name}}.', 'SKD/{sequence}/RT02/2026', 'active', 48),
    ('22222222-2222-4222-8222-222222223202', '22222222-2222-4222-8222-222222222222', 'Surat Pengantar Usaha',
     '["KTP","Kartu Keluarga","Keterangan usaha"]'::jsonb, '{"nama_usaha":{"type":"string","label":"Nama usaha"}}'::jsonb,
     'Surat pengantar usaha untuk {{full_name}}.', 'SKU/{sequence}/RT02/2026', 'active', 72)
ON CONFLICT (organization_id, name) DO NOTHING;

INSERT INTO letter_requests (
    id, organization_id, requester_user_id, resident_id, letter_type_id, request_number, form_data, status,
    resident_note, submitted_at, processed_by, sla_due_at
) VALUES (
    '22222222-2222-4222-8222-222222223203', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222305',
    '22222222-2222-4222-8222-222222222501', '22222222-2222-4222-8222-222222223201', 'REQ-RT02-2026-0001',
    '{"keperluan":"Administrasi sekolah anak"}'::jsonb, 'under_review', 'Mohon diproses sebelum hari Jumat.',
    now() - INTERVAL '1 day', '22222222-2222-4222-8222-222222222302', now() + INTERVAL '1 day'
) ON CONFLICT (organization_id, request_number) DO NOTHING;

INSERT INTO complaint_categories (
    id, organization_id, code, name, status, target_response_hours, target_resolution_hours, target_reporter_confirmation_hours
) VALUES (
    '22222222-2222-4222-8222-222222223301', '22222222-2222-4222-8222-222222222222', 'lampu_jalan', 'Lampu Jalan',
    'active', 24, 72, 48
) ON CONFLICT (organization_id, code) DO NOTHING;

INSERT INTO complaints (
    id, organization_id, reporter_user_id, ticket_number, complaint_category_id, title, description,
    location_description, priority, status, assigned_to, response_due_at, responded_at, resolution_due_at
) VALUES (
    '22222222-2222-4222-8222-222222223302', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222306',
    'ADU-RT02-2026-0001', '22222222-2222-4222-8222-222222223301', 'Lampu jalan di Blok A mati',
    'Lampu jalan di depan Blok A Nomor 02 mati sejak dua malam terakhir.', 'Depan Blok A Nomor 02',
    'high', 'in_progress', '22222222-2222-4222-8222-222222222304', now() - INTERVAL '20 hours', now() - INTERVAL '18 hours', now() + INTERVAL '2 days'
) ON CONFLICT (organization_id, ticket_number) DO NOTHING;

INSERT INTO complaint_comments (id, organization_id, complaint_id, author_user_id, body, is_internal) VALUES
    ('22222222-2222-4222-8222-222222223311', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222223302', '22222222-2222-4222-8222-222222222304', 'Petugas sudah dijadwalkan memeriksa jaringan lampu sore ini.', false)
ON CONFLICT (organization_id, id) DO NOTHING;

INSERT INTO complaint_events (id, organization_id, complaint_id, actor_user_id, event_type, data) VALUES
    ('22222222-2222-4222-8222-222222223312', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222223302', '22222222-2222-4222-8222-222222222306', 'submitted', '{"status":"new"}'::jsonb),
    ('22222222-2222-4222-8222-222222223313', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222223302', '22222222-2222-4222-8222-222222222304', 'assigned', '{"status":"in_progress"}'::jsonb)
ON CONFLICT DO NOTHING;

INSERT INTO notification_preferences (
    organization_id, user_id, in_app_enabled, email_enabled, whatsapp_enabled, due_reminder_enabled
) VALUES
    ('22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222305', true, true, false, true),
    ('22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222306', true, true, false, true)
ON CONFLICT (organization_id, user_id) DO NOTHING;

INSERT INTO notifications (id, organization_id, user_id, type, title, body, reference_type, reference_id, read_at) VALUES
    ('22222222-2222-4222-8222-222222223401', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222305', 'letter_request', 'Pengajuan surat sedang diproses', 'Pengajuan Surat Keterangan Domisili telah diterima sekretaris.', 'letter_request', '22222222-2222-4222-8222-222222223203', NULL),
    ('22222222-2222-4222-8222-222222223402', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222306', 'complaint', 'Aduan sedang ditangani', 'Petugas telah dijadwalkan memeriksa lampu jalan.', 'complaint', '22222222-2222-4222-8222-222222223302', NULL),
    ('22222222-2222-4222-8222-222222223403', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222304', 'announcement', 'Pengumuman kerja bakti', 'Kerja bakti lingkungan akan dilaksanakan lima hari lagi.', 'announcement', '22222222-2222-4222-8222-222222223101', now())
ON CONFLICT (organization_id, id) DO NOTHING;

INSERT INTO invoice_generation_runs (
    id, organization_id, due_type_id, period_start, period_end, due_date, run_key, status,
    total_targeted, total_created, total_skipped, started_at, completed_at, created_by
) VALUES (
    '22222222-2222-4222-8222-222222223501', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222222702',
    date_trunc('month', CURRENT_DATE)::date, (date_trunc('month', CURRENT_DATE + INTERVAL '1 month')::date - 1),
    CURRENT_DATE + 3, 'demo-run-current-month', 'completed', 3, 3, 0, now() - INTERVAL '7 days', now() - INTERVAL '7 days' + INTERVAL '2 seconds',
    '22222222-2222-4222-8222-222222222303'
) ON CONFLICT (organization_id, run_key) DO NOTHING;

INSERT INTO cash_publication_policies (organization_id, is_public, public_until)
VALUES ('22222222-2222-4222-8222-222222222222', true, CURRENT_DATE + 31)
ON CONFLICT (organization_id) DO UPDATE SET is_public = EXCLUDED.is_public, public_until = EXCLUDED.public_until;

INSERT INTO cash_publication_categories (organization_id, cash_category_id) VALUES
    ('22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222223001'),
    ('22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222223002'),
    ('22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222223003')
ON CONFLICT DO NOTHING;

INSERT INTO public_cash_summaries (
    id, organization_id, period_start, period_end, total_income, total_expense, ending_balance, published_at
) VALUES (
    '22222222-2222-4222-8222-222222223601', '22222222-2222-4222-8222-222222222222',
    date_trunc('month', CURRENT_DATE - INTERVAL '1 month')::date, (date_trunc('month', CURRENT_DATE)::date - 1),
    375000, 150000, 225000, now() - INTERVAL '1 day'
) ON CONFLICT (organization_id, period_start, period_end) DO NOTHING;

INSERT INTO public_cash_summary_categories (
    public_cash_summary_id, organization_id, cash_category_id, category_name, transaction_type, total_amount
) VALUES
    ('22222222-2222-4222-8222-222222223601', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222223001', 'Penerimaan Iuran', 'income', 375000),
    ('22222222-2222-4222-8222-222222223601', '22222222-2222-4222-8222-222222222222', '22222222-2222-4222-8222-222222223002', 'Operasional Lingkungan', 'expense', 150000)
ON CONFLICT DO NOTHING;
`)
	return err
}

func seedUsers(ctx context.Context, tx pgx.Tx, passwordHash string, linkResidents bool) error {
	users := []demoUser{
		{ketuaID, "", "ketua.rt02@example.test", "ketua_rt"},
		{sekretarisID, "", "sekretaris.rt02@example.test", "sekretaris"},
		{bendaharaID, "", "bendahara.rt02@example.test", "bendahara"},
		{pengurusID, "", "pengurus.rt02@example.test", "pengurus"},
		{wargaBudiID, budiID, "budi.demo@example.test", "warga"},
		{wargaSitiID, sitiID, "siti.demo@example.test", "warga"},
	}

	for _, user := range users {
		_, err := tx.Exec(ctx, `
INSERT INTO users (id, organization_id, resident_id, email, password_hash, status, failed_login_count, last_login_at)
VALUES ($1, $2, CASE WHEN $6 THEN NULLIF($3, '')::uuid END, $4, $5, 'active', 0, now() - INTERVAL '1 day')
ON CONFLICT (organization_id, lower(email)) WHERE email IS NOT NULL
DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    status = 'active',
    failed_login_count = 0,
    locked_until = NULL,
    resident_id = CASE WHEN $6 THEN EXCLUDED.resident_id ELSE users.resident_id END`,
			user.id, orgID, user.residentID, user.email, passwordHash, linkResidents,
		)
		if err != nil {
			return fmt.Errorf("upsert %s: %w", user.email, err)
		}

		_, err = tx.Exec(ctx, `
INSERT INTO user_roles (user_id, role_id)
SELECT $1, id FROM roles WHERE organization_id IS NULL AND code = $2
ON CONFLICT DO NOTHING`, user.id, user.role)
		if err != nil {
			return fmt.Errorf("assign %s role: %w", user.role, err)
		}
	}
	return nil
}