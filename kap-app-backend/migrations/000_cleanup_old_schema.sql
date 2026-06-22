-- =============================================================================
-- KAP-APP: Eski Veritabanı Temizleme (Cleanup) Scripti
-- Migration  : 000_cleanup_old_schema
-- Açıklama   : Önceki denemelerden kalan tüm tabloları ve fonksiyonları siler.
-- Dikkat     : Bu işlem tablo içeriklerindeki tüm verileri sıfırlar!
-- =============================================================================

BEGIN;

-- 1. Tetikleyicileri (Triggers) Kaldır
DROP TRIGGER IF EXISTS trg_profiles_updated_at ON public.profiles CASCADE;
DROP TRIGGER IF EXISTS trg_tenants_updated_at ON public.tenants CASCADE;
DROP TRIGGER IF EXISTS trg_tenant_memberships_updated_at ON public.tenant_memberships CASCADE;
DROP TRIGGER IF EXISTS trg_products_updated_at ON public.products CASCADE;
DROP TRIGGER IF EXISTS trg_on_auth_user_created ON auth.users CASCADE;

-- 2. Görünümleri (Views) Kaldır
DROP VIEW IF EXISTS public.v_user_tenant_roles CASCADE;

-- 3. Tabloları Kaldır (CASCADE ile ilişkisel bağımlılıklarıyla birlikte silinir)
DROP TABLE IF EXISTS public.products CASCADE;
DROP TABLE IF EXISTS public.tenant_memberships CASCADE;
DROP TABLE IF EXISTS public.tenants CASCADE;
DROP TABLE IF EXISTS public.profiles CASCADE;

-- 4. Fonksiyonları Kaldır
DROP FUNCTION IF EXISTS public.handle_updated_at() CASCADE;
DROP FUNCTION IF EXISTS public.generate_unique_slug(int) CASCADE;
DROP FUNCTION IF EXISTS public.handle_new_user() CASCADE;
DROP FUNCTION IF EXISTS public.get_current_user_tenant_ids() CASCADE;
DROP FUNCTION IF EXISTS public.get_current_user_role_in_tenant(uuid) CASCADE;

COMMIT;
