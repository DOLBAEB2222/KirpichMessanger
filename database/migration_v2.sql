-- Migration v2: Wiki, Code Snippets, Temporary Roles, RSS Aggregator
-- Run this migration to add new features to existing database

-- =============================================
-- Add Code Message Type
-- =============================================
-- Note: In PostgreSQL, you can't add values to an ENUM directly
-- This requires recreating the type. For production, use a migration tool.

-- Alternative: Add new message_type as varchar and check in application
-- For now, we'll handle this in the application layer

-- =============================================
-- Wiki Pages for Channels
-- =============================================

CREATE TABLE wiki_pages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_by_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES wiki_pages(id) ON DELETE SET NULL,
    is_published BOOLEAN DEFAULT TRUE,
    "order" INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(channel_id, slug)
);

CREATE INDEX idx_wiki_pages_channel ON wiki_pages(channel_id);
CREATE INDEX idx_wiki_pages_parent ON wiki_pages(parent_id);
CREATE INDEX idx_wiki_pages_published ON wiki_pages(is_published) WHERE is_published = TRUE;

-- =============================================
-- Code Snippets
-- =============================================

CREATE TYPE code_language AS ENUM (
    'javascript', 'typescript', 'python', 'go', 'java',
    'c', 'cpp', 'rust', 'php', 'ruby', 'sql',
    'html', 'css', 'bash', 'json', 'xml', 'markdown', 'other'
);

CREATE TABLE code_snippets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL UNIQUE REFERENCES messages(id) ON DELETE CASCADE,
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    language code_language NOT NULL,
    code TEXT NOT NULL,
    file_name VARCHAR(255),
    created_by_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_code_snippets_message ON code_snippets(message_id);
CREATE INDEX idx_code_snippets_chat ON code_snippets(chat_id);
CREATE INDEX idx_code_snippets_language ON code_snippets(language);

-- =============================================
-- Temporary Roles
-- =============================================

CREATE TYPE temp_role_type AS ENUM ('moderator', 'admin', 'custom');
CREATE TYPE temp_role_target AS ENUM ('chat', 'channel');

CREATE TABLE temp_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    target_id UUID NOT NULL,
    target_type temp_role_target NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_type temp_role_type NOT NULL,
    custom_name VARCHAR(255),
    permissions TEXT[] NOT NULL,
    granted_by_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_temp_roles_target ON temp_roles(target_id, target_type);
CREATE INDEX idx_temp_roles_user ON temp_roles(user_id);
CREATE INDEX idx_temp_roles_expires ON temp_roles(expires_at) WHERE is_active = TRUE;
CREATE INDEX idx_temp_roles_active ON temp_roles(is_active, user_id);

-- =============================================
-- RSS Feeds and Items
-- =============================================

CREATE TABLE rss_feeds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id UUID NOT NULL UNIQUE REFERENCES channels(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    icon_url TEXT,
    added_by_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT TRUE,
    last_fetched TIMESTAMP,
    fetch_error TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_rss_feeds_channel ON rss_feeds(channel_id);
CREATE INDEX idx_rss_feeds_active ON rss_feeds(is_active);
CREATE INDEX idx_rss_feeds_last_fetch ON rss_feeds(last_fetched DESC);

CREATE TABLE rss_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    feed_id UUID NOT NULL REFERENCES rss_feeds(id) ON DELETE CASCADE,
    guid VARCHAR(500) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    content TEXT,
    link TEXT NOT NULL,
    author VARCHAR(255),
    category VARCHAR(255),
    published_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    UNIQUE(feed_id, guid)
);

CREATE INDEX idx_rss_items_feed ON rss_items(feed_id);
CREATE INDEX idx_rss_items_guid ON rss_items(guid);
CREATE INDEX idx_rss_items_published ON rss_items(published_at DESC);
CREATE INDEX idx_rss_items_message ON rss_items(message_id);

-- =============================================
-- Update Triggers
-- =============================================

CREATE TRIGGER update_wiki_pages_updated_at BEFORE UPDATE ON wiki_pages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_code_snippets_updated_at BEFORE UPDATE ON code_snippets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rss_feeds_updated_at BEFORE UPDATE ON rss_feeds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================
-- Comments
-- =============================================

COMMENT ON TABLE wiki_pages IS 'Wiki pages for channels documentation';
COMMENT ON TABLE code_snippets IS 'Code snippets attached to messages with syntax highlighting';
COMMENT ON TABLE temp_roles IS 'Temporary role assignments with expiration for chats/channels';
COMMENT ON TABLE rss_feeds IS 'RSS feeds subscribed to channels';
COMMENT ON TABLE rss_items IS 'RSS feed items that can be posted as messages';

-- =============================================
-- Cleanup
-- =============================================

VACUUM ANALYZE;

SELECT 'Migration v2 completed successfully!' AS status;
