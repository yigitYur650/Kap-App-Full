-- Migration: 002_add_product_unit
-- Açıklama: products tablosuna unit (birim) kolonu eklenmesi.

BEGIN;

ALTER TABLE public.products
    ADD COLUMN IF NOT EXISTS unit TEXT;

COMMIT;
