-- ===========================================
-- Table: user_rewards
-- ===========================================

CREATE TABLE user_rewards (
                            id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4 (),
                            user_id UUID NOT NULL,     -- Reference to users.id (logical only)
                            reward_id UUID NOT NULL,   -- Reference to rewards.id (logical only)
                            reason TEXT,               -- e.g. "complete_set", "leaderboard_top10"
                            granted_at TIMESTAMPTZ NOT NULL DEFAULT now (),
  -- audit & lifecycle
                            created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
                            updated_at TIMESTAMPTZ,
                            deleted_at TIMESTAMPTZ
);

-- Automatically update 'updated_at' on row modifications
CREATE TRIGGER trigger_user_rewards_updated_at
  BEFORE UPDATE ON user_rewards
  FOR EACH ROW
  EXECUTE FUNCTION trigger_updated_at ();

-- Helpful indexes
CREATE INDEX idx_user_rewards_user_id
  ON user_rewards USING btree (user_id)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_user_rewards_reward_id
  ON user_rewards USING btree (reward_id)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_user_rewards_user_id_reason
  ON user_rewards USING btree (user_id, reason)
  WHERE deleted_at IS NULL;