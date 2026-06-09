-- migrations/000002_create_social_accounts.down.sql

-- Must drop table BEFORE dropping the type it depends on
DROP TABLE IF EXISTS social_accounts;
DROP TYPE  IF EXISTS social_platform;