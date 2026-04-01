-- Tables for Master Data
CREATE TABLE master_genders (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    code VARCHAR(255) UNIQUE,
    name TEXT,
    icon TEXT,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE master_relationship_types (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    code VARCHAR(255) UNIQUE,
    name TEXT,
    icon TEXT,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE master_interests (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    name TEXT,
    icon TEXT,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE master_languages (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    code VARCHAR(255) UNIQUE,
    name TEXT,
    icon TEXT,
    is_active BOOLEAN DEFAULT TRUE
);

-- Core Tables
CREATE TABLE entities (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    type VARCHAR(20) NOT NULL
);
CREATE INDEX idx_entities_type ON entities(type);
CREATE INDEX idx_entities_created_at ON entities(created_at);
CREATE INDEX idx_entities_updated_at ON entities(updated_at);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    entity_id UUID NOT NULL REFERENCES entities(id),
    email VARCHAR(255) UNIQUE,
    password_hash TEXT,
    full_name TEXT,
    date_of_birth TIMESTAMP WITH TIME ZONE,
    height_cm INT,
    bio TEXT,
    location_city TEXT,
    location_country TEXT,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    gender_id UUID REFERENCES master_genders(id),
    relationship_type_id UUID REFERENCES master_relationship_types(id),
    is_premium BOOLEAN DEFAULT FALSE,
    last_active_at TIMESTAMP WITH TIME ZONE,
    status TEXT,
    age INT,
    swipe_count_today INT DEFAULT 0,
    verified_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_users_entity_id ON users(entity_id);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_date_of_birth ON users(date_of_birth);
CREATE INDEX idx_users_height_cm ON users(height_cm);
CREATE INDEX idx_users_latitude ON users(latitude);
CREATE INDEX idx_users_longitude ON users(longitude);
CREATE INDEX idx_users_gender_id ON users(gender_id);
CREATE INDEX idx_users_relationship_type_id ON users(relationship_type_id);
CREATE INDEX idx_users_is_premium ON users(is_premium);
CREATE INDEX idx_users_last_active_at ON users(last_active_at);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_age ON users(age);
CREATE INDEX idx_users_verified_at ON users(verified_at);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_updated_at ON users(updated_at);

CREATE TABLE groups (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    entity_id UUID NOT NULL REFERENCES entities(id),
    name VARCHAR(255) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id)
);
CREATE INDEX idx_groups_entity_id ON groups(entity_id);
CREATE INDEX idx_groups_created_by ON groups(created_by);
CREATE INDEX idx_groups_created_at ON groups(created_at);
CREATE INDEX idx_groups_updated_at ON groups(updated_at);

CREATE TABLE group_members (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_admin BOOLEAN DEFAULT FALSE
);
CREATE INDEX idx_group_members_group_id ON group_members(group_id);
CREATE INDEX idx_group_members_user_id ON group_members(user_id);
CREATE UNIQUE INDEX idx_group_pair ON group_members(group_id, user_id);
CREATE INDEX idx_group_members_created_at ON group_members(created_at);
CREATE INDEX idx_group_members_updated_at ON group_members(updated_at);

-- Interaction Tables
CREATE TABLE swipes (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    swiper_entity_id UUID NOT NULL REFERENCES entities(id),
    swiped_entity_id UUID NOT NULL REFERENCES entities(id),
    direction VARCHAR(20) NOT NULL,
    is_boosted BOOLEAN DEFAULT FALSE,
    ranking_score DOUBLE PRECISION DEFAULT 0,
    priority_score INT DEFAULT 0,
    processed_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_swipes_swiper_entity_id ON swipes(swiper_entity_id);
CREATE INDEX idx_swipes_swiped_entity_id ON swipes(swiped_entity_id);
CREATE UNIQUE INDEX idx_entity_swipe ON swipes(swiper_entity_id, swiped_entity_id);
CREATE INDEX idx_swipes_direction ON swipes(direction);
CREATE INDEX idx_swipes_is_boosted ON swipes(is_boosted);
CREATE INDEX idx_swipes_ranking_score ON swipes(ranking_score);
CREATE INDEX idx_swipes_priority_score ON swipes(priority_score);
CREATE INDEX idx_swipes_processed_at ON swipes(processed_at);
CREATE INDEX idx_swipes_created_at ON swipes(created_at);
CREATE INDEX idx_swipes_updated_at ON swipes(updated_at);

CREATE TABLE matches (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    entity1_id UUID NOT NULL REFERENCES entities(id),
    entity2_id UUID NOT NULL REFERENCES entities(id)
);
CREATE INDEX idx_matches_entity1_id ON matches(entity1_id);
CREATE INDEX idx_matches_entity2_id ON matches(entity2_id);
CREATE UNIQUE INDEX idx_entity_pair ON matches(entity1_id, entity2_id);
CREATE INDEX idx_matches_created_at ON matches(created_at);
CREATE INDEX idx_matches_updated_at ON matches(updated_at);

CREATE TABLE entity_boosts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    entity_id UUID NOT NULL REFERENCES entities(id),
    started_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_entity_boosts_entity_id ON entity_boosts(entity_id);
CREATE INDEX idx_entity_boosts_started_at ON entity_boosts(started_at);
CREATE INDEX idx_entity_boosts_expires_at ON entity_boosts(expires_at);
CREATE INDEX idx_entity_boosts_created_at ON entity_boosts(created_at);
CREATE INDEX idx_entity_boosts_updated_at ON entity_boosts(updated_at);

CREATE TABLE entity_impressions (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    viewer_entity_id UUID NOT NULL REFERENCES entities(id),
    shown_entity_id UUID NOT NULL REFERENCES entities(id),
    shown_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_entity_impressions_viewer_entity_id ON entity_impressions(viewer_entity_id);
CREATE INDEX idx_entity_impressions_shown_entity_id ON entity_impressions(shown_entity_id);
CREATE INDEX idx_entity_impressions_shown_at ON entity_impressions(shown_at);
CREATE INDEX idx_entity_impressions_created_at ON entity_impressions(created_at);
CREATE INDEX idx_entity_impressions_updated_at ON entity_impressions(updated_at);

CREATE TABLE entity_unmatches (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    swiper_entity_id UUID NOT NULL REFERENCES entities(id),
    target_entity_id UUID NOT NULL REFERENCES entities(id)
);
CREATE INDEX idx_entity_unmatches_swiper_entity_id ON entity_unmatches(swiper_entity_id);
CREATE INDEX idx_entity_unmatches_target_entity_id ON entity_unmatches(target_entity_id);
CREATE INDEX idx_entity_unmatches_created_at ON entity_unmatches(created_at);
CREATE INDEX idx_entity_unmatches_updated_at ON entity_unmatches(updated_at);

-- Chat Tables
CREATE TABLE conversations (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    type VARCHAR(20) DEFAULT 'direct',
    entity_id UUID REFERENCES matches(id),
    last_message_id UUID,
    last_message_at TIMESTAMP WITH TIME ZONE,
    visible_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_conversations_deleted_at ON conversations(deleted_at);
CREATE INDEX idx_conversations_type ON conversations(type);
CREATE INDEX idx_conversations_entity_id ON conversations(entity_id);
CREATE INDEX idx_conversations_last_message_id ON conversations(last_message_id);
CREATE INDEX idx_conversations_last_message_at ON conversations(last_message_at);
CREATE INDEX idx_conversations_visible_at ON conversations(visible_at);
CREATE INDEX idx_conversations_created_at ON conversations(created_at);
CREATE INDEX idx_conversations_updated_at ON conversations(updated_at);

CREATE TABLE conversation_participants (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_message_id UUID
);
CREATE INDEX idx_conversation_participants_conversation_id ON conversation_participants(conversation_id);
CREATE INDEX idx_conversation_participants_user_id ON conversation_participants(user_id);
CREATE UNIQUE INDEX idx_conv_user ON conversation_participants(conversation_id, user_id);
CREATE INDEX idx_conversation_participants_last_read_message_id ON conversation_participants(last_read_message_id);
CREATE INDEX idx_conversation_participants_created_at ON conversation_participants(created_at);
CREATE INDEX idx_conversation_participants_updated_at ON conversation_participants(updated_at);

CREATE TABLE messages (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'sent',
    content TEXT,
    reply_to_id UUID REFERENCES messages(id),
    metadata JSONB
);
CREATE INDEX idx_messages_deleted_at ON messages(deleted_at);
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_conv_created ON messages(conversation_id, created_at);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_status ON messages(status);
CREATE INDEX idx_messages_reply_to_id ON messages(reply_to_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_updated_at ON messages(updated_at);

CREATE TABLE message_reads (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    conversation_id UUID NOT NULL REFERENCES conversations(id)
);
CREATE INDEX idx_message_reads_message_id ON message_reads(message_id);
CREATE INDEX idx_message_reads_user_id ON message_reads(user_id);
CREATE INDEX idx_message_reads_conversation_id ON message_reads(conversation_id);
CREATE INDEX idx_message_reads_created_at ON message_reads(created_at);
CREATE INDEX idx_message_reads_updated_at ON message_reads(updated_at);

CREATE TABLE group_invites (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    inviter_id UUID NOT NULL REFERENCES users(id),
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    used_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_group_invites_group_id ON group_invites(group_id);
CREATE INDEX idx_group_invites_inviter_id ON group_invites(inviter_id);
CREATE INDEX idx_group_invites_expires_at ON group_invites(expires_at);
CREATE INDEX idx_group_invites_used_at ON group_invites(used_at);
CREATE INDEX idx_group_invites_created_at ON group_invites(created_at);
CREATE INDEX idx_group_invites_updated_at ON group_invites(updated_at);

CREATE TABLE user_presences (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    is_online BOOLEAN,
    last_seen_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_user_presences_is_online ON user_presences(is_online);
CREATE INDEX idx_user_presences_last_seen_at ON user_presences(last_seen_at);
CREATE INDEX idx_user_presences_created_at ON user_presences(created_at);
CREATE INDEX idx_user_presences_updated_at ON user_presences(updated_at);

-- Auth & Device Tables
CREATE TABLE auth_providers (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(255),
    provider_user_id VARCHAR(255)
);
CREATE INDEX idx_auth_providers_user_id ON auth_providers(user_id);
CREATE UNIQUE INDEX idx_provider_user ON auth_providers(provider, provider_user_id);
CREATE INDEX idx_auth_providers_created_at ON auth_providers(created_at);
CREATE INDEX idx_auth_providers_updated_at ON auth_providers(updated_at);

CREATE TABLE photos (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT,
    is_main BOOLEAN,
    sort_order INT
);
CREATE INDEX idx_photos_user_id ON photos(user_id);
CREATE INDEX idx_photos_created_at ON photos(created_at);
CREATE INDEX idx_photos_updated_at ON photos(updated_at);

CREATE TABLE devices (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255),
    device_name TEXT,
    device_model TEXT,
    os_version TEXT,
    app_version TEXT,
    fcm_token TEXT,
    last_ip TEXT,
    last_login TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN
);
CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE UNIQUE INDEX idx_device_user ON devices(user_id, device_id);
CREATE INDEX idx_devices_created_at ON devices(created_at);
CREATE INDEX idx_devices_updated_at ON devices(updated_at);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_device_id ON refresh_tokens(device_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_revoked_at ON refresh_tokens(revoked_at);
CREATE INDEX idx_refresh_tokens_created_at ON refresh_tokens(created_at);
CREATE INDEX idx_refresh_tokens_updated_at ON refresh_tokens(updated_at);

-- Pivot Tables (M2M)
CREATE TABLE user_interested_genders (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    gender_id UUID NOT NULL REFERENCES master_genders(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, gender_id)
);
CREATE INDEX idx_user_interested_genders_user_id ON user_interested_genders(user_id);
CREATE INDEX idx_user_interested_genders_gender_id ON user_interested_genders(gender_id);

CREATE TABLE user_interests (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interest_id UUID NOT NULL REFERENCES master_interests(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, interest_id)
);
CREATE INDEX idx_user_interests_user_id ON user_interests(user_id);
CREATE INDEX idx_user_interests_interest_id ON user_interests(interest_id);

CREATE TABLE user_languages (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES master_languages(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, language_id)
);
CREATE INDEX idx_user_languages_user_id ON user_languages(user_id);
CREATE INDEX idx_user_languages_language_id ON user_languages(language_id);

-- Config & Monetization
CREATE TABLE app_configs (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    key VARCHAR(255) UNIQUE,
    value JSONB,
    description TEXT
);
CREATE INDEX idx_app_configs_created_at ON app_configs(created_at);
CREATE INDEX idx_app_configs_updated_at ON app_configs(updated_at);

CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT TRUE
);
CREATE INDEX idx_subscription_plans_created_at ON subscription_plans(created_at);
CREATE INDEX idx_subscription_plans_updated_at ON subscription_plans(updated_at);

CREATE TABLE subscription_prices (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id) ON DELETE CASCADE,
    duration_type TEXT NOT NULL,
    price NUMERIC NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    external_slug VARCHAR(255) UNIQUE
);
CREATE INDEX idx_subscription_prices_plan_id ON subscription_prices(plan_id);
CREATE INDEX idx_subscription_prices_duration_type ON subscription_prices(duration_type);
CREATE INDEX idx_subscription_prices_created_at ON subscription_prices(created_at);
CREATE INDEX idx_subscription_prices_updated_at ON subscription_prices(updated_at);

CREATE TABLE subscription_plan_features (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id) ON DELETE CASCADE,
    feature_key VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    category VARCHAR(50) DEFAULT 'General',
    icon TEXT,
    display_title TEXT,
    is_consumable BOOLEAN DEFAULT FALSE,
    amount INT DEFAULT 0
);
CREATE INDEX idx_subscription_plan_features_plan_id ON subscription_plan_features(plan_id);
CREATE UNIQUE INDEX idx_plan_feature ON subscription_plan_features(plan_id, feature_key);
CREATE INDEX idx_subscription_plan_features_created_at ON subscription_plan_features(created_at);
CREATE INDEX idx_subscription_plan_features_updated_at ON subscription_plan_features(updated_at);

CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    started_at TIMESTAMP WITH TIME ZONE,
    expired_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE
);
CREATE INDEX idx_user_subscriptions_deleted_at ON user_subscriptions(deleted_at);
CREATE UNIQUE INDEX idx_user_subs ON user_subscriptions(user_id);
CREATE INDEX idx_user_subscriptions_user_id ON user_subscriptions(user_id);
CREATE INDEX idx_user_subscriptions_plan_id ON user_subscriptions(plan_id);
CREATE INDEX idx_user_subscriptions_started_at ON user_subscriptions(started_at);
CREATE INDEX idx_user_subscriptions_expired_at ON user_subscriptions(expired_at);
CREATE INDEX idx_user_subscriptions_is_active ON user_subscriptions(is_active);
CREATE INDEX idx_user_subscriptions_created_at ON user_subscriptions(created_at);
CREATE INDEX idx_user_subscriptions_updated_at ON user_subscriptions(updated_at);

CREATE TABLE user_consumables (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_type VARCHAR(255) NOT NULL,
    amount INT DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX idx_user_cons ON user_consumables(user_id, item_type);
CREATE INDEX idx_user_consumables_user_id ON user_consumables(user_id);
CREATE INDEX idx_user_consumables_item_type ON user_consumables(item_type);
CREATE INDEX idx_user_consumables_last_used_at ON user_consumables(last_used_at);
CREATE INDEX idx_user_consumables_created_at ON user_consumables(created_at);
CREATE INDEX idx_user_consumables_updated_at ON user_consumables(updated_at);

CREATE TABLE consumable_packages (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    name TEXT,
    item_type VARCHAR(255) NOT NULL,
    amount INT NOT NULL,
    price NUMERIC NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);
CREATE INDEX idx_consumable_packages_item_type ON consumable_packages(item_type);
CREATE INDEX idx_consumable_packages_created_at ON consumable_packages(created_at);
CREATE INDEX idx_consumable_packages_updated_at ON consumable_packages(updated_at);

-- Background Jobs
CREATE TABLE jobs (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    type VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    payload JSONB,
    reference_id UUID,
    reference_type VARCHAR(255),
    source VARCHAR(255),
    error_message TEXT,
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    processed_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_jobs_type ON jobs(type);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_reference_id ON jobs(reference_id);
CREATE INDEX idx_jobs_reference_type ON jobs(reference_type);
CREATE INDEX idx_jobs_source ON jobs(source);
CREATE INDEX idx_jobs_processed_at ON jobs(processed_at);
CREATE INDEX idx_jobs_created_at ON jobs(created_at);
CREATE INDEX idx_jobs_updated_at ON jobs(updated_at);

-- Notifications
CREATE TABLE notification_settings (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    type VARCHAR(255) NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT,
    is_enable BOOLEAN DEFAULT TRUE
);
CREATE INDEX idx_notification_settings_created_at ON notification_settings(created_at);
CREATE INDEX idx_notification_settings_updated_at ON notification_settings(updated_at);

CREATE TABLE user_notification_settings (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_setting_id UUID NOT NULL REFERENCES notification_settings(id) ON DELETE CASCADE,
    is_enable BOOLEAN DEFAULT TRUE
);
CREATE INDEX idx_user_notification_settings_user_id ON user_notification_settings(user_id);
CREATE INDEX idx_user_notification_settings_notification_setting_id ON user_notification_settings(notification_setting_id);
CREATE UNIQUE INDEX idx_user_notif_setting ON user_notification_settings(user_id, notification_setting_id);
CREATE INDEX idx_user_notification_settings_created_at ON user_notification_settings(created_at);
CREATE INDEX idx_user_notification_settings_updated_at ON user_notification_settings(updated_at);
