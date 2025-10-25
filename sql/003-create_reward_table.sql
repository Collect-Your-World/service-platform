-- ===========================================
-- Table: rewards
-- ===========================================

CREATE TABLE rewards (
                       id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4 (),
                       code TEXT NOT NULL UNIQUE,             -- Unique reward code
                       type TEXT NOT NULL,                    -- badge, coins, sweepstakes, cash
                       name TEXT NOT NULL,                    -- Reward name
                       description TEXT,                      -- Description
                       value_int INT,                         -- Coins or tickets
                       meta_jsonb JSONB,                      -- Extra metadata
  -- audit & lifecycle
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
                       updated_at TIMESTAMPTZ,
                       deleted_at TIMESTAMPTZ                 -- soft delete
);

-- Automatically update 'updated_at' on row modifications
CREATE TRIGGER trigger_rewards_updated_at
  BEFORE UPDATE ON rewards
  FOR EACH ROW
  EXECUTE FUNCTION trigger_updated_at ();

-- Helpful indexes
CREATE INDEX idx_rewards_code
  ON rewards USING btree (code)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_rewards_type
  ON rewards USING btree (type)
  WHERE deleted_at IS NULL;