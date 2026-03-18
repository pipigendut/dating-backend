-- 1. SUBSCRIPTION SYSTEM (FLEXIBLE CONFIG)

CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE, -- plus, premium, ultimate
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE subscription_plan_features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES subscription_plans(id) ON DELETE CASCADE,
    feature_key VARCHAR(100) NOT NULL, -- see_likes, priority_likes, unlimited_likes, free_boost_monthly
    value JSONB NOT NULL, -- e.g. true, 10, "high"
    UNIQUE(plan_id, feature_key)
);

CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expired_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for fast lookup of active subscription
CREATE INDEX idx_user_subscriptions_active ON user_subscriptions(user_id) WHERE is_active = true AND expired_at > NOW();

-- 2. CONSUMABLE FEATURES

CREATE TABLE user_consumables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL, -- boost, crush
    remaining INTEGER DEFAULT 0,
    expired_at TIMESTAMP WITH TIME ZONE, -- nullable for permanent items
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_user_consumables_lookup ON user_consumables(user_id, type);

-- 3. USER FIELDS EXTENSION (already in users table, but listed here for documentation)
-- ALTER TABLE users ADD COLUMN boosts_until TIMESTAMP WITH TIME ZONE;
-- ALTER TABLE users ADD COLUMN swipe_count_today INTEGER DEFAULT 0;
-- ALTER TABLE users ADD COLUMN last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- 4. RANKING & IMPRESSION CONTROL

CREATE TABLE user_impressions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    viewer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shown_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shown_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index to quickly filter out already shown users
CREATE INDEX idx_viewer_shown_users ON user_impressions(viewer_id, shown_user_id);
-- Index to cleanup old impressions (retention policy)
CREATE INDEX idx_shown_at ON user_impressions(shown_at);

-- 5. APP CONFIG TABLE

CREATE TABLE app_configs (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- INITIAL CONFIG SEED
INSERT INTO app_configs (key, value, description) VALUES
('boost_multiplier', '5', 'Multiplier for ranking score during boost'),
('crush_priority_score', '100', 'Priority score for crush swipes'),
('swipe_limit_free', '50', 'Daily swipe limit for free users'),
('swipe_limit_premium', '1000', 'Daily swipe limit for premium users'),
('delay_free_minutes', '60', 'Delay in minutes before showing a like to free users'),
('delay_premium_minutes', '10', 'Delay in minutes before showing a like to premium users');

-- 6. SWIPE EXTENSION
-- ALTER TABLE swipes ADD COLUMN priority_score INTEGER DEFAULT 0;
-- ALTER TABLE swipes ADD COLUMN processed_at TIMESTAMP WITH TIME ZONE; -- For anti-fast match delay
