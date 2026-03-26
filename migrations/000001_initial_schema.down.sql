-- 000001_initial_schema.down.sql

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS user_presences;
DROP TABLE IF EXISTS message_reads;
ALTER TABLE conversations DROP CONSTRAINT IF EXISTS fk_last_message;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversation_participants;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS consumable_packages;
DROP TABLE IF EXISTS user_consumables;
DROP TABLE IF EXISTS user_subscriptions;
DROP TABLE IF EXISTS subscription_plan_features;
DROP TABLE IF EXISTS subscription_prices;
DROP TABLE IF EXISTS subscription_plans;
DROP TABLE IF EXISTS app_configs;
DROP TABLE IF EXISTS user_boosts;
DROP TABLE IF EXISTS user_impressions;
DROP TABLE IF EXISTS unmatches;
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS swipes;
DROP TABLE IF EXISTS user_languages;
DROP TABLE IF EXISTS user_interests;
DROP TABLE IF EXISTS user_interested_genders;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS auth_providers;
DROP TABLE IF EXISTS photos;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS master_languages;
DROP TABLE IF EXISTS master_interests;
DROP TABLE IF EXISTS master_relationship_types;
DROP TABLE IF EXISTS master_genders;

-- Optional: Drop UUID extension (usually better to keep if other things use it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
