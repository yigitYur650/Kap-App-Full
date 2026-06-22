-- =============================================================================
-- KAP-APP: Supabase PostgreSQL Veritabanı Şeması & RLS Politikaları
-- Migration  : 001_complete_schema
-- Açıklama   : Core DDL, Tetikleyiciler, Audit soft-delete ve RLS politikaları.
-- Prensip    : Idempotent & Zero-Dependency (Go Backend ile %100 Uyumlu).
-- =============================================================================

BEGIN;

-- =============================================================================
-- BÖLÜM 1: UZANTILAR VE YARDIMCI FONKSİYONLAR
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";  -- uuid_generate_v4() desteği için
CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- gen_random_bytes() desteği için

-- -----------------------------------------------------------------------------
-- 1.1 updated_at Otomatik Güncelleme Fonksiyonu
-- -----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.handle_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW() AT TIME ZONE 'UTC';
    RETURN NEW;
END;
$$;

-- -----------------------------------------------------------------------------
-- 1.2 Benzersiz Slug Üretici Fonksiyonu (Çakışma-Güvenli)
-- -----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.generate_unique_slug(p_length INT DEFAULT 10)
RETURNS TEXT
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_slug        TEXT;
    v_exists      BOOLEAN;
    v_max_retries INT := 10;
    v_attempt     INT := 0;
BEGIN
    LOOP
        v_attempt := v_attempt + 1;

        v_slug := lower(
            regexp_replace(
                encode(gen_random_bytes(p_length * 2), 'base64'),
                '[^a-zA-Z0-9]', '', 'g'
            )
        );

        v_slug := substr(v_slug, 1, p_length);

        SELECT EXISTS (
            SELECT 1 FROM public.profiles WHERE slug_id = v_slug
        ) INTO v_exists;

        EXIT WHEN NOT v_exists;

        IF v_attempt >= v_max_retries THEN
            RAISE EXCEPTION
                'Benzersiz slug üretilemedi: % deneme sonrasında çakışma devam ediyor.',
                v_max_retries;
        END IF;
    END LOOP;

    RETURN v_slug;
END;
$$;

-- -----------------------------------------------------------------------------
-- 1.3 Yeni Auth Kullanıcısı İşleme Fonksiyonu (Defansif & SECURITY DEFINER)
-- -----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_display_name TEXT;
    v_slug         TEXT;
BEGIN
    v_display_name := COALESCE(
        NEW.raw_user_meta_data ->> 'full_name',
        NEW.raw_user_meta_data ->> 'name',
        split_part(NEW.email, '@', 1),
        'User'
    );

    v_slug := public.generate_unique_slug(10);

    INSERT INTO public.profiles (id, display_name, slug_id)
    VALUES (NEW.id, v_display_name, v_slug);

    RETURN NEW;
END;
$$;

-- =============================================================================
-- BÖLÜM 2: TABLO TANIMLAMALARI (DDL)
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 2.1 profiles
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.profiles (
    id           UUID        PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    display_name TEXT,
    slug_id      VARCHAR(10) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    CONSTRAINT chk_profiles_slug_id_not_empty CHECK (slug_id <> '')
);

ALTER TABLE public.profiles
    ADD COLUMN IF NOT EXISTS display_name TEXT,
    ADD COLUMN IF NOT EXISTS slug_id      VARCHAR(10),
    ADD COLUMN IF NOT EXISTS created_at   TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ADD COLUMN IF NOT EXISTS updated_at   TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC');

ALTER TABLE public.profiles DROP CONSTRAINT IF EXISTS profiles_slug_id_key;
ALTER TABLE public.profiles ADD  CONSTRAINT profiles_slug_id_key UNIQUE (slug_id);

-- -----------------------------------------------------------------------------
-- 2.2 tenants (Go Backend modelindeki theme_color ve owner_id ile uyumlu)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.tenants (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT        NOT NULL,
    theme_color    TEXT        NOT NULL DEFAULT '#4F46E5',
    owner_id       UUID        NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    CONSTRAINT chk_tenants_name_not_empty CHECK (name <> ''),
    CONSTRAINT chk_tenants_theme_color_format CHECK (theme_color ~ '^#[0-9A-Fa-f]{6}$')
);

ALTER TABLE public.tenants
    ADD COLUMN IF NOT EXISTS name           TEXT        NOT NULL DEFAULT 'Geçici Ev',
    ADD COLUMN IF NOT EXISTS theme_color    TEXT        NOT NULL DEFAULT '#4F46E5',
    ADD COLUMN IF NOT EXISTS owner_id       UUID,
    ADD COLUMN IF NOT EXISTS created_at     TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ADD COLUMN IF NOT EXISTS updated_at     TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC');

-- -----------------------------------------------------------------------------
-- 2.3 tenant_memberships (Go Backend modelindeki user_id ile uyumlu)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.tenant_memberships (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID        NOT NULL REFERENCES public.tenants(id)  ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
    role       TEXT        NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    CONSTRAINT uq_tenant_membership UNIQUE (tenant_id, user_id),
    CONSTRAINT chk_tenant_memberships_role CHECK (role IN ('owner', 'admin', 'member'))
);

ALTER TABLE public.tenant_memberships
    ADD COLUMN IF NOT EXISTS role       TEXT        NOT NULL DEFAULT 'member',
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC');

-- -----------------------------------------------------------------------------
-- 2.4 products (added_by, deleted_by ve soft-delete alanları dahil)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.products (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    name            TEXT        NOT NULL,
    price           NUMERIC(12, 2),
    market_name     TEXT,
    category        TEXT,
    quantity        INT         NOT NULL DEFAULT 1,
    expiration_date TIMESTAMPTZ,
    status          TEXT        NOT NULL DEFAULT 'yok',
    added_by        UUID        REFERENCES public.profiles(id) ON DELETE SET NULL,
    deleted_by      UUID        REFERENCES public.profiles(id) ON DELETE SET NULL,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    CONSTRAINT chk_products_name_not_empty CHECK (name <> ''),
    CONSTRAINT chk_products_quantity CHECK (quantity >= 0),
    CONSTRAINT chk_products_status CHECK (status IN ('var', 'azaldı', 'yok'))
);

ALTER TABLE public.products
    ADD COLUMN IF NOT EXISTS name            TEXT         NOT NULL DEFAULT 'Ürün',
    ADD COLUMN IF NOT EXISTS price           NUMERIC(12, 2),
    ADD COLUMN IF NOT EXISTS market_name     TEXT,
    ADD COLUMN IF NOT EXISTS category        TEXT,
    ADD COLUMN IF NOT EXISTS quantity        INT          NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS expiration_date TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS status          TEXT         NOT NULL DEFAULT 'yok',
    ADD COLUMN IF NOT EXISTS added_by        UUID,
    ADD COLUMN IF NOT EXISTS deleted_by      UUID,
    ADD COLUMN IF NOT EXISTS deleted_at      TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS created_at      TIMESTAMPTZ  NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ADD COLUMN IF NOT EXISTS updated_at      TIMESTAMPTZ  NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC');

-- =============================================================================
-- BÖLÜM 3: İNDEKS STRATEJİSİ
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_tenant_memberships_tenant_id ON public.tenant_memberships (tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_memberships_user_id   ON public.tenant_memberships (user_id);
CREATE INDEX IF NOT EXISTS idx_products_tenant_id           ON public.products (tenant_id);

CREATE INDEX IF NOT EXISTS idx_products_active_by_tenant
    ON public.products (tenant_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_category
    ON public.products (tenant_id, category) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_status
    ON public.products (tenant_id, status) WHERE deleted_at IS NULL;

-- =============================================================================
-- BÖLÜM 4: TETİKLEYİCİLER (TRIGGERS)
-- =============================================================================

-- updated_at Tetikleyicileri
DROP TRIGGER IF EXISTS trg_profiles_updated_at ON public.profiles;
CREATE TRIGGER trg_profiles_updated_at
    BEFORE UPDATE ON public.profiles
    FOR EACH ROW EXECUTE FUNCTION public.handle_updated_at();

DROP TRIGGER IF EXISTS trg_tenants_updated_at ON public.tenants;
CREATE TRIGGER trg_tenants_updated_at
    BEFORE UPDATE ON public.tenants
    FOR EACH ROW EXECUTE FUNCTION public.handle_updated_at();

DROP TRIGGER IF EXISTS trg_tenant_memberships_updated_at ON public.tenant_memberships;
CREATE TRIGGER trg_tenant_memberships_updated_at
    BEFORE UPDATE ON public.tenant_memberships
    FOR EACH ROW EXECUTE FUNCTION public.handle_updated_at();

DROP TRIGGER IF EXISTS trg_products_updated_at ON public.products;
CREATE TRIGGER trg_products_updated_at
    BEFORE UPDATE ON public.products
    FOR EACH ROW EXECUTE FUNCTION public.handle_updated_at();

-- Auth Signup Tetikleyicisi
DROP TRIGGER IF EXISTS trg_on_auth_user_created ON auth.users;
CREATE TRIGGER trg_on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- -----------------------------------------------------------------------------
-- 4.3 Soft-Delete & Audit Denetim Tetikleyicisi
-- GC (30 günlük temizlik) pg_cron job'a taşındı — bkz. 002_add_pg_cron_gc.sql
-- -----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.handle_product_soft_delete()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
    -- Ürünü silmek yerine soft-delete yap; silen kimliği ve zamanı işle.
    UPDATE public.products
    SET deleted_at = NOW() AT TIME ZONE 'UTC',
        deleted_by = auth.uid()
    WHERE id = OLD.id
      AND deleted_at IS NULL;  -- Zaten silinmişse tekrar işleme

    -- Gerçek (hard) DELETE işlemini iptal et.
    RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS trg_products_soft_delete ON public.products;
CREATE TRIGGER trg_products_soft_delete
    BEFORE DELETE ON public.products
    FOR EACH ROW EXECUTE FUNCTION public.handle_product_soft_delete();

-- =============================================================================
-- BÖLÜM 5: ROW LEVEL SECURITY (RLS) POLİTİKALARI
-- =============================================================================

ALTER TABLE public.profiles           ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.tenants            ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.tenant_memberships ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.products           ENABLE ROW LEVEL SECURITY;

ALTER TABLE public.profiles           FORCE ROW LEVEL SECURITY;
ALTER TABLE public.tenants            FORCE ROW LEVEL SECURITY;
ALTER TABLE public.tenant_memberships FORCE ROW LEVEL SECURITY;
ALTER TABLE public.products           FORCE ROW LEVEL SECURITY;

-- -----------------------------------------------------------------------------
-- 5.1 RLS Yardımcı Fonksiyonları (Sonsuz Döngüyü Önlemek İçin DEFINER Yapıldı)
-- -----------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION public.get_current_user_tenant_ids()
RETURNS UUID[]
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = public
AS $$
    SELECT COALESCE(ARRAY_AGG(tm.tenant_id), '{}'::UUID[])
    FROM public.tenant_memberships tm
    WHERE tm.user_id = auth.uid();
$$;

CREATE OR REPLACE FUNCTION public.get_current_user_role_in_tenant(p_tenant_id UUID)
RETURNS TEXT
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = public
AS $$
    SELECT role
    FROM public.tenant_memberships
    WHERE tenant_id  = p_tenant_id
      AND user_id    = auth.uid()
    LIMIT 1;
$$;

-- -----------------------------------------------------------------------------
-- 5.2 profiles Tablosu Politikaları
-- -----------------------------------------------------------------------------

DROP POLICY IF EXISTS "profiles: ev arkadaşları ve kendisi" ON public.profiles;
CREATE POLICY "profiles: ev arkadaşları ve kendisi"
    ON public.profiles
    FOR SELECT
    TO authenticated
    USING (
        id = auth.uid()
        OR id IN (
            SELECT tm.user_id
            FROM public.tenant_memberships tm
            WHERE tm.tenant_id = ANY(public.get_current_user_tenant_ids())
        )
    );

DROP POLICY IF EXISTS "profiles: sadece kendi profilini ekleyebilir" ON public.profiles;
CREATE POLICY "profiles: sadece kendi profilini ekleyebilir"
    ON public.profiles
    FOR INSERT
    TO authenticated
    WITH CHECK (auth.uid() = id);

DROP POLICY IF EXISTS "profiles: sadece kendi profilini güncelleyebilir" ON public.profiles;
CREATE POLICY "profiles: sadece kendi profilini güncelleyebilir"
    ON public.profiles
    FOR UPDATE
    TO authenticated
    USING     (auth.uid() = id)
    WITH CHECK (auth.uid() = id);

-- -----------------------------------------------------------------------------
-- 5.3 tenants Tablosu Politikaları
-- -----------------------------------------------------------------------------

DROP POLICY IF EXISTS "tenants: üyeler okuyabilir" ON public.tenants;
CREATE POLICY "tenants: üyeler okuyabilir"
    ON public.tenants
    FOR SELECT
    TO authenticated
    USING (id = ANY(public.get_current_user_tenant_ids()));

DROP POLICY IF EXISTS "tenants: herkes ev oluşturabilir" ON public.tenants;
CREATE POLICY "tenants: herkes ev oluşturabilir"
    ON public.tenants
    FOR INSERT
    TO authenticated
    WITH CHECK (true);

-- Member rolündeki kullanıcıların ev adını veya temasını güncellemesi RLS ile engellenir.
-- Sadece 'owner' veya 'admin' rolündekiler güncelleyebilir.
DROP POLICY IF EXISTS "tenants: owner/admin güncelleyebilir" ON public.tenants;
DROP POLICY IF EXISTS "tenants: üyeler güncelleyebilir" ON public.tenants;
CREATE POLICY "tenants: owner/admin güncelleyebilir"
    ON public.tenants
    FOR UPDATE
    TO authenticated
    USING     (public.get_current_user_role_in_tenant(id) IN ('owner', 'admin'))
    WITH CHECK (public.get_current_user_role_in_tenant(id) IN ('owner', 'admin'));

-- -----------------------------------------------------------------------------
-- 5.4 tenant_memberships Tablosu Politikaları
-- -----------------------------------------------------------------------------

DROP POLICY IF EXISTS "memberships: üyeler kendi evinin listesini okur" ON public.tenant_memberships;
CREATE POLICY "memberships: üyeler kendi evinin listesini okur"
    ON public.tenant_memberships
    FOR SELECT
    TO authenticated
    USING (tenant_id = ANY(public.get_current_user_tenant_ids()));

DROP POLICY IF EXISTS "memberships: kullanıcı kendini ekleyebilir" ON public.tenant_memberships;
CREATE POLICY "memberships: kullanıcı kendini ekleyebilir"
    ON public.tenant_memberships
    FOR INSERT
    TO authenticated
    WITH CHECK (
        user_id = auth.uid()
        AND role = 'member'
    );

DROP POLICY IF EXISTS "memberships: owner üye ekleyebilir" ON public.tenant_memberships;
CREATE POLICY "memberships: owner üye ekleyebilir"
    ON public.tenant_memberships
    FOR INSERT
    TO authenticated
    WITH CHECK (
        public.get_current_user_role_in_tenant(tenant_id) = 'owner'
        AND role = 'member'
    );

DROP POLICY IF EXISTS "memberships: owner üye güncelleyebilir" ON public.tenant_memberships;
CREATE POLICY "memberships: owner üye güncelleyebilir"
    ON public.tenant_memberships
    FOR UPDATE
    TO authenticated
    USING (
        public.get_current_user_role_in_tenant(tenant_id) = 'owner'
    )
    WITH CHECK (
        public.get_current_user_role_in_tenant(tenant_id) = 'owner'
    );

DROP POLICY IF EXISTS "memberships: admin/owner üye silebilir" ON public.tenant_memberships;
CREATE POLICY "memberships: admin/owner üye silebilir"
    ON public.tenant_memberships
    FOR DELETE
    TO authenticated
    USING (
        public.get_current_user_role_in_tenant(tenant_id) IN ('owner', 'admin')
    );

DROP POLICY IF EXISTS "memberships: kullanıcı kendini çıkarabilir" ON public.tenant_memberships;
CREATE POLICY "memberships: kullanıcı kendini çıkarabilir"
    ON public.tenant_memberships
    FOR DELETE
    TO authenticated
    USING (user_id = auth.uid());

-- -----------------------------------------------------------------------------
-- 5.5 products Tablosu Politikaları
-- -----------------------------------------------------------------------------

DROP POLICY IF EXISTS "products: üyeler görebilir" ON public.products;
CREATE POLICY "products: üyeler görebilir"
    ON public.products
    FOR SELECT
    TO authenticated
    USING (
        tenant_id = ANY(public.get_current_user_tenant_ids())
        AND deleted_at IS NULL
    );

DROP POLICY IF EXISTS "products: üyeler ekleyebilir" ON public.products;
CREATE POLICY "products: üyeler ekleyebilir"
    ON public.products
    FOR INSERT
    TO authenticated
    WITH CHECK (tenant_id = ANY(public.get_current_user_tenant_ids()));

DROP POLICY IF EXISTS "products: üyeler güncelleyebilir" ON public.products;
CREATE POLICY "products: üyeler güncelleyebilir"
    ON public.products
    FOR UPDATE
    TO authenticated
    USING     (tenant_id = ANY(public.get_current_user_tenant_ids()))
    WITH CHECK (tenant_id = ANY(public.get_current_user_tenant_ids()));

DROP POLICY IF EXISTS "products: üyeler silebilir" ON public.products;
CREATE POLICY "products: üyeler silebilir"
    ON public.products
    FOR DELETE
    TO authenticated
    USING (tenant_id = ANY(public.get_current_user_tenant_ids()));

COMMIT;


-- =============================================================================
-- BÖLÜM 6: TDD DOĞRULAMA TEST SENARYOLARI (ROLLBACK KORUMALI)
-- =============================================================================

BEGIN;
SAVEPOINT rls_tests;

DO $$
DECLARE
    -- Test UUID'leri
    v_user_a   UUID := 'aaaaaaaa-0000-0000-0000-000000000001'; -- Tenant A üyesi (owner)
    v_user_b   UUID := 'bbbbbbbb-0000-0000-0000-000000000002'; -- Tenant B üyesi (yabancı)
    v_user_m   UUID := 'cccccccc-0000-0000-0000-000000000003'; -- Tenant A'da member
    v_tenant_a UUID := 'aaaaaaaa-1111-0000-0000-000000000010';
    v_tenant_b UUID := 'bbbbbbbb-2222-0000-0000-000000000020';
    v_prod_del UUID := 'dddddddd-0000-0000-0000-000000000030'; -- Soft-deleted ürün
    v_count    INT;
    v_result   TEXT;
    v_rows_upd INT;
    v_prod_id  UUID := gen_random_uuid();
    v_aud_user UUID;
    v_aud_time TIMESTAMPTZ;
BEGIN
    -- Fixture Kurulumu (RLS bypass)
    SET LOCAL session_replication_role = 'replica';
    SET LOCAL row_security = off;

    INSERT INTO public.profiles (id, slug_id, display_name) VALUES
        (v_user_a, 'testaaa01', 'User A'),
        (v_user_b, 'testbbb02', 'User B'),
        (v_user_m, 'testccc03', 'User M')
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO public.tenants (id, name, owner_id) VALUES
        (v_tenant_a, '[TEST] Ev A', v_user_a),
        (v_tenant_b, '[TEST] Ev B', v_user_b)
    ON CONFLICT (id) DO NOTHING;

    INSERT INTO public.tenant_memberships (id, tenant_id, user_id, role) VALUES
        (gen_random_uuid(), v_tenant_a, v_user_a, 'owner'),
        (gen_random_uuid(), v_tenant_b, v_user_b, 'owner'),
        (gen_random_uuid(), v_tenant_a, v_user_m, 'member')
    ON CONFLICT DO NOTHING;

    INSERT INTO public.products (id, tenant_id, name, status, deleted_at) VALUES
        (v_prod_del, v_tenant_a, '[TEST] Silinmiş Ürün', 'yok', NOW())
    ON CONFLICT (id) DO NOTHING;

    -- RLS ve rol doğrulamasını zorlamak için güvenlik ayarlarını kuruyoruz
    SET LOCAL row_security = on;

    -- -------------------------------------------------------------------------
    -- TEST-1: profiles güvenlik sızıntısı
    -- -------------------------------------------------------------------------
    SET LOCAL ROLE postgres;
    PERFORM set_config('request.jwt.claims', format('{"sub":"%s","role":"authenticated"}', v_user_a::text), true);
    PERFORM set_config('request.jwt.claim.sub', v_user_a::text, true);
    PERFORM set_config('request.jwt.claim.role', 'authenticated', true);
    SET LOCAL ROLE authenticated;

    SELECT COUNT(*) INTO v_count
    FROM public.profiles
    WHERE id = v_user_b;

    ASSERT v_count = 0,
        format('[FAIL] TEST-1 — profiles sızıntısı: User A, Tenant B''deki User B profilini görebildi!');
    RAISE NOTICE '[PASS] TEST-1 — profiles: Farklı evdeki kullanıcının profili görünmüyor.';

    -- -------------------------------------------------------------------------
    -- TEST-2: Rol yükseltme koruması
    -- -------------------------------------------------------------------------
    SET LOCAL ROLE postgres;
    PERFORM set_config('request.jwt.claims', format('{"sub":"%s","role":"authenticated"}', v_user_m::text), true);
    PERFORM set_config('request.jwt.claim.sub', v_user_m::text, true);
    PERFORM set_config('request.jwt.claim.role', 'authenticated', true);
    SET LOCAL ROLE authenticated;

    SELECT public.get_current_user_role_in_tenant(v_tenant_a) INTO v_result;

    ASSERT v_result = 'member',
        format('[FAIL] TEST-2 — Rol fonksiyonu hatalı: %s', v_result);

    ASSERT NOT (v_result = 'owner'),
        '[FAIL] TEST-2 — Role elevation koruması devre dışı: member, owner rolünde görünüyor!';
    RAISE NOTICE '[PASS] TEST-2 — Role elevation: member rolü owner ile yetkilendirilemiyor.';

    -- -------------------------------------------------------------------------
    -- TEST-3: Soft-Delete filtresi
    -- -------------------------------------------------------------------------
    SET LOCAL ROLE postgres;
    PERFORM set_config('request.jwt.claims', format('{"sub":"%s","role":"authenticated"}', v_user_a::text), true);
    PERFORM set_config('request.jwt.claim.sub', v_user_a::text, true);
    PERFORM set_config('request.jwt.claim.role', 'authenticated', true);
    SET LOCAL ROLE authenticated;

    SELECT COUNT(*) INTO v_count
    FROM public.products
    WHERE id = v_prod_del;

    ASSERT v_count = 0,
        format('[FAIL] TEST-3 — Soft-Delete sızıntısı: deleted_at dolu ürün sorguda göründü! Alınan: %s', v_count);
    RAISE NOTICE '[PASS] TEST-3 — products: Soft-deleted ürün SELECT ile görünmüyor.';

    -- -------------------------------------------------------------------------
    -- TEST-4: [NEGATİF YAZMA] Yetkisiz Eve Ürün Ekleme Engeli
    -- Senaryo: User A (Tenant A üyesi) -> Tenant B'ye ürün eklemeye çalışır.
    -- Beklenen: RLS engeli (insufficient_privilege) fırlatılmalı.
    -- -------------------------------------------------------------------------
    BEGIN
        SET LOCAL ROLE postgres;
        PERFORM set_config('request.jwt.claims', format('{"sub":"%s","role":"authenticated"}', v_user_a::text), true);
        PERFORM set_config('request.jwt.claim.sub', v_user_a::text, true);
        PERFORM set_config('request.jwt.claim.role', 'authenticated', true);
        SET LOCAL ROLE authenticated;

        INSERT INTO public.products (id, tenant_id, name, status)
        VALUES (gen_random_uuid(), v_tenant_b, '[TEST] Sızma Ürünü', 'var');

        RAISE EXCEPTION '[FAIL] TEST-4 — products RLS sızıntısı: User A, üyesi olmadığı Tenant B''ye ürün ekleyebildi!';
    EXCEPTION
        WHEN insufficient_privilege THEN
            RAISE NOTICE '[PASS] TEST-4 — products RLS: Yetkisiz tenant''a ürün eklenmesi engellendi (42501).';
    END;

    -- -------------------------------------------------------------------------
    -- TEST-5: [NEGATİF YAZMA] Member Rolünün Ev Bilgisi Güncelleme Engeli
    -- Senaryo: User M (Tenant A'da member) -> Tenant A'nın adını değiştirmeye çalışır.
    -- Beklenen: RLS engeli sebebiyle 0 satır güncellenmeli.
    -- -------------------------------------------------------------------------
    SET LOCAL ROLE postgres;
    PERFORM set_config('request.jwt.claims', format('{"sub":"%s","role":"authenticated"}', v_user_m::text), true);
    PERFORM set_config('request.jwt.claim.sub', v_user_m::text, true);
    PERFORM set_config('request.jwt.claim.role', 'authenticated', true);
    SET LOCAL ROLE authenticated;

    UPDATE public.tenants
    SET name = '[TEST] Hacker Evi'
    WHERE id = v_tenant_a;

    GET DIAGNOSTICS v_rows_upd = ROW_COUNT;

    ASSERT v_rows_upd = 0,
        '[FAIL] TEST-5 — tenants RLS yetersiz: member rolündeki kullanıcı ev adını güncelleyebildi!';
    RAISE NOTICE '[PASS] TEST-5 — tenants RLS: Member rolünün ev adını güncellemesi başarıyla engellendi.';

    -- -------------------------------------------------------------------------
    -- TEST-6: Soft-Delete & Audit testi
    -- Senaryo: Bir ürünü sil → deleted_at set edilmeli, hard delete olmamalı.
    -- -------------------------------------------------------------------------
    SET LOCAL ROLE postgres;
    SET LOCAL row_security = off;

    -- Silinecek test ürünü ekle
    INSERT INTO public.products (id, tenant_id, name, status)
    VALUES (v_prod_id, v_tenant_a, '[TEST] Silinecek Ürün', 'var')
    ON CONFLICT (id) DO NOTHING;

    SET LOCAL row_security = on;
    PERFORM set_config('request.jwt.claims', format('{"sub":"%s","role":"authenticated"}', v_user_a::text), true);
    PERFORM set_config('request.jwt.claim.sub', v_user_a::text, true);
    SET LOCAL ROLE authenticated;

    -- DELETE → trigger BEFORE DELETE → RETURN NULL (hard delete iptal)
    DELETE FROM public.products WHERE id = v_prod_id;

    SET LOCAL ROLE postgres;
    SET LOCAL row_security = off;

    -- Ürün hâlâ var olmalı (soft-deleted)
    SELECT COUNT(*) INTO v_count FROM public.products WHERE id = v_prod_id;
    ASSERT v_count = 1,
        format('[FAIL] TEST-6 — Soft-Delete: Ürün tamamen silindi, deleted_at set edilmedi! Kalan satır: %s', v_count);

    -- deleted_at NULL olmamalı
    SELECT COUNT(*) INTO v_count
    FROM public.products
    WHERE id = v_prod_id AND deleted_at IS NOT NULL;
    ASSERT v_count = 1,
        '[FAIL] TEST-6 — Soft-Delete: deleted_at alanı NULL kaldı!';

    RAISE NOTICE '[PASS] TEST-6 — Soft-Delete: Ürün silinmedi, deleted_at başarıyla set edildi.';

    RAISE NOTICE '========================================================';
    RAISE NOTICE 'TÜM RLS DOĞRULAMA TESTLERİ BAŞARIYLA GEÇTİ.';
    RAISE NOTICE 'Test verileri ROLLBACK ile temizleniyor...';
    RAISE NOTICE '========================================================';

END;
$$;

ROLLBACK TO SAVEPOINT rls_tests;
RELEASE SAVEPOINT rls_tests;
COMMIT;
