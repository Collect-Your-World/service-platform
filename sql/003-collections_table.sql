-- =====================================================================
-- 003-collections_table.sql
-- Schema for collections and collection items
-- =====================================================================

-- =====================
-- 1. Collections
-- =====================
CREATE TABLE
    collections (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
        name VARCHAR(100) NOT NULL,
        slug VARCHAR(150) NOT NULL,
        description TEXT,
        collection_type VARCHAR(30) NOT NULL DEFAULT 'THEME', -- THEME / LOCATION / COUNTRY / GLOBAL
        reward_amount BIGINT DEFAULT 0, -- can represent tokens, coins, or gift card cents
        reward_currency VARCHAR(20) NOT NULL, -- "COIN", "SPIN", etc.
        is_enabled BOOLEAN DEFAULT TRUE,
        created_at TIMESTAMPTZ DEFAULT NOW (),
        updated_at TIMESTAMPTZ,
        deleted_at TIMESTAMPTZ -- soft delete
    );

-- User: Global top rewards
CREATE UNIQUE INDEX unique_idx_collections_slug ON collections (slug)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_collections_enabled_reward ON collections (is_enabled, reward_amount DESC, reward_currency)
WHERE
    deleted_at IS NULL;

-- =====================
-- 2. Collection Items
-- =====================
CREATE TABLE
    collection_items (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
        collection_id UUID NOT NULL,
        item_id UUID NOT NULL,
        image_url TEXT,
        created_at TIMESTAMPTZ DEFAULT NOW (),
        updated_at TIMESTAMPTZ,
        deleted_at TIMESTAMPTZ -- soft delete
    );

CREATE UNIQUE INDEX unique_idx_collection_items_collection_item ON collection_items (collection_id, item_id)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_collection_items_collection_id ON collection_items USING btree (collection_id)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_collection_items_item_id ON collection_items USING btree (item_id)
WHERE
    deleted_at IS NULL;

ALTER TABLE collection_items
    ADD CONSTRAINT fk_collection_items_collection_id FOREIGN KEY (collection_id) REFERENCES collections (id) ON DELETE CASCADE;

CREATE TRIGGER trigger_collections_updated_at BEFORE
UPDATE ON collections FOR EACH ROW EXECUTE FUNCTION trigger_updated_at ();

CREATE TRIGGER trigger_collection_items_updated_at BEFORE
UPDATE ON collection_items FOR EACH ROW EXECUTE FUNCTION trigger_updated_at ();
