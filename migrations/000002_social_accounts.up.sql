-- migrations/000002_create_social_accounts.up.sql

-- Create the platform enum
-- Adding a new platform later = ALTER TYPE social_platform ADD VALUE 'tiktok'
CREATE TYPE social_platform AS ENUM (
    'linkedin',
    'instagram',
    'x',
    'facebook'
);

CREATE TABLE social_accounts (
    id               UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id          UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform         social_platform NOT NULL,
    platform_user_id VARCHAR(255)  NOT NULL,
    name             VARCHAR(255),
    email            VARCHAR(255),
    access_token     TEXT          NOT NULL,
    token_expires_at TIMESTAMPTZ   NOT NULL,
    created_at       TIMESTAMPTZ   DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   DEFAULT NOW(),

    -- one user can only connect one account per platform
    -- so a user can have linkedin + instagram but not two linkedins
    CONSTRAINT unique_user_platform UNIQUE (user_id, platform)
);