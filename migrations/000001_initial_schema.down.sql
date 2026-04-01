-- Drop Notification Tables
DROP TABLE IF EXISTS user_notification_settings;
DROP TABLE IF EXISTS notification_settings;

-- Drop Job Tables
DROP TABLE IF EXISTS jobs;

-- Drop Config & Monetization Tables
DROP TABLE IF EXISTS consumable_packages;
DROP TABLE IF EXISTS user_consumables;
DROP TABLE IF EXISTS user_subscriptions;
DROP TABLE IF EXISTS subscription_plan_features;
DROP TABLE IF EXISTS subscription_prices;
DROP TABLE IF EXISTS subscription_plans;
DROP TABLE IF EXISTS app_configs;

-- Drop Pivot Tables
DROP TABLE IF EXISTS user_languages;
DROP TABLE IF EXISTS user_interests;
DROP TABLE IF EXISTS user_interested_genders;

-- Drop Auth & Device Tables
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS photos;
DROP TABLE IF EXISTS auth_providers;
DROP TABLE IF EXISTS user_presences;

-- Drop Chat Tables
DROP TABLE IF EXISTS group_invites;
DROP TABLE IF EXISTS message_reads;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversation_participants;
DROP TABLE IF EXISTS conversations;

-- Drop Interaction Tables
DROP TABLE IF EXISTS entity_unmatches;
DROP TABLE IF EXISTS entity_impressions;
DROP TABLE IF EXISTS entity_boosts;
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS swipes;

-- Drop Core Tables
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS entities;

-- Drop Master Data Tables
DROP TABLE IF EXISTS master_languages;
DROP TABLE IF EXISTS master_interests;
DROP TABLE IF EXISTS master_relationship_types;
DROP TABLE IF EXISTS master_genders;
