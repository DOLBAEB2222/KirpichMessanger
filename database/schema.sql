-- Messenger Application Database Schema
-- PostgreSQL 16
-- Optimized for Intel i3-2120, 4GB RAM

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create ENUM types
CREATE TYPE chat_type AS ENUM ('dm', 'group');
CREATE TYPE message_type AS ENUM ('text', 'image', 'video', 'audio', 'file');
CREATE TYPE member_role AS ENUM ('admin', 'member');
CREATE TYPE subscription_type AS ENUM ('premium_monthly', 'premium_yearly');
CREATE TYPE subscription_status AS ENUM ('active', 'expired', 'cancelled');
CREATE TYPE payment_status AS ENUM ('completed_stub', 'pending', 'failed', 'completed', 'refunded');

-- =============================================
-- Users Table
-- =============================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    username VARCHAR(50) UNIQUE,
    avatar_url TEXT,
    bio TEXT,
    is_premium BOOLEAN DEFAULT FALSE,
    last_seen_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for users
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_is_premium ON users(is_premium);

-- =============================================
-- Chats Table (DM and Groups)
-- =============================================
CREATE TABLE chats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255),
    type chat_type NOT NULL,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    avatar_url TEXT,
    description TEXT,
    member_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_message_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for chats
CREATE INDEX idx_chats_owner ON chats(owner_id);
CREATE INDEX idx_chats_type ON chats(type);
CREATE INDEX idx_chats_updated ON chats(updated_at DESC);

-- =============================================
-- Chat Members (Many-to-Many)
-- =============================================
CREATE TABLE chat_members (
    chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role member_role DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT NOW(),
    last_read_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (chat_id, user_id)
);

-- Indexes for chat_members
CREATE INDEX idx_chat_members_user ON chat_members(user_id);
CREATE INDEX idx_chat_members_chat ON chat_members(chat_id);

-- =============================================
-- Messages Table
-- =============================================
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sender_id UUID REFERENCES users(id) ON DELETE SET NULL,
    chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    message_type message_type DEFAULT 'text',
    media_url TEXT,
    media_size BIGINT,
    reply_to_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    is_edited BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for messages (optimized for pagination)
CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_messages_chat_id ON messages(chat_id) WHERE is_deleted = FALSE;
CREATE INDEX idx_messages_reply_to ON messages(reply_to_id) WHERE reply_to_id IS NOT NULL;

-- =============================================
-- Channels Table (One-to-Many Broadcast)
-- =============================================
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    owner_id UUID REFERENCES users(id) ON DELETE CASCADE,
    description TEXT,
    avatar_url TEXT,
    subscriber_count INTEGER DEFAULT 0,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for channels
CREATE INDEX idx_channels_owner ON channels(owner_id);
CREATE INDEX idx_channels_public ON channels(is_public) WHERE is_public = TRUE;

-- =============================================
-- Channel Subscribers
-- =============================================
CREATE TABLE channel_subscribers (
    channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    subscribed_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (channel_id, user_id)
);

-- Indexes for channel_subscribers
CREATE INDEX idx_channel_subscribers_user ON channel_subscribers(user_id);
CREATE INDEX idx_channel_subscribers_channel ON channel_subscribers(channel_id);

-- =============================================
-- Subscriptions Table (Premium Features)
-- =============================================
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    type subscription_type NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    auto_renew BOOLEAN DEFAULT FALSE,
    status subscription_status DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for subscriptions
CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_user_status ON subscriptions(user_id, status);
CREATE INDEX idx_subscriptions_end_date ON subscriptions(end_date) WHERE status = 'active';

-- =============================================
-- Payment Logs (MVP: Stub Implementation)
-- =============================================
CREATE TABLE payment_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    amount DECIMAL(10, 2) NOT NULL,
    subscription_type VARCHAR(50) NOT NULL,
    status payment_status DEFAULT 'pending',
    payment_method VARCHAR(50) DEFAULT 'stub',
    transaction_id VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for payment_logs
CREATE INDEX idx_payment_logs_user ON payment_logs(user_id);
CREATE INDEX idx_payment_logs_status ON payment_logs(status);
CREATE INDEX idx_payment_logs_created ON payment_logs(created_at DESC);

-- =============================================
-- Session Tokens (JWT Refresh Tokens)
-- =============================================
CREATE TABLE session_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    device_info JSONB,
    ip_address INET,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for session_tokens
CREATE INDEX idx_session_tokens_user ON session_tokens(user_id);
CREATE INDEX idx_session_tokens_hash ON session_tokens(token_hash);
CREATE INDEX idx_session_tokens_expires ON session_tokens(expires_at);

-- =============================================
-- User Contacts (Address Book)
-- =============================================
CREATE TABLE contacts (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    contact_id UUID REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(255),
    added_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, contact_id),
    CHECK (user_id != contact_id)
);

-- Indexes for contacts
CREATE INDEX idx_contacts_user ON contacts(user_id);

-- =============================================
-- Blocked Users
-- =============================================
CREATE TABLE blocked_users (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    blocked_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    blocked_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, blocked_user_id),
    CHECK (user_id != blocked_user_id)
);

-- Indexes for blocked_users
CREATE INDEX idx_blocked_users_user ON blocked_users(user_id);

-- =============================================
-- Media Files (Track uploads for cleanup)
-- =============================================
CREATE TABLE media_files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    file_path TEXT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for media_files
CREATE INDEX idx_media_files_user ON media_files(user_id);
CREATE INDEX idx_media_files_message ON media_files(message_id);
CREATE INDEX idx_media_files_created ON media_files(created_at);

-- =============================================
-- Functions and Triggers
-- =============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to relevant tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chats_updated_at BEFORE UPDATE ON chats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_channels_updated_at BEFORE UPDATE ON channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to update chat member count
CREATE OR REPLACE FUNCTION update_chat_member_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE chats SET member_count = member_count + 1 WHERE id = NEW.chat_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE chats SET member_count = member_count - 1 WHERE id = OLD.chat_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_chat_member_count_trigger
AFTER INSERT OR DELETE ON chat_members
FOR EACH ROW EXECUTE FUNCTION update_chat_member_count();

-- Function to update channel subscriber count
CREATE OR REPLACE FUNCTION update_channel_subscriber_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE channels SET subscriber_count = subscriber_count + 1 WHERE id = NEW.channel_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE channels SET subscriber_count = subscriber_count - 1 WHERE id = OLD.channel_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_channel_subscriber_count_trigger
AFTER INSERT OR DELETE ON channel_subscribers
FOR EACH ROW EXECUTE FUNCTION update_channel_subscriber_count();

-- Function to update chat last_message_at
CREATE OR REPLACE FUNCTION update_chat_last_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE chats SET last_message_at = NEW.created_at WHERE id = NEW.chat_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_chat_last_message_trigger
AFTER INSERT ON messages
FOR EACH ROW EXECUTE FUNCTION update_chat_last_message();

-- =============================================
-- Views for Common Queries
-- =============================================

-- View for user profile with premium status
CREATE VIEW user_profiles AS
SELECT 
    u.id,
    u.phone,
    u.email,
    u.username,
    u.avatar_url,
    u.bio,
    u.is_premium,
    u.last_seen_at,
    u.created_at,
    CASE 
        WHEN s.status = 'active' AND s.end_date >= CURRENT_DATE THEN TRUE
        ELSE FALSE
    END AS has_active_subscription,
    s.end_date AS subscription_end_date
FROM users u
LEFT JOIN subscriptions s ON u.id = s.user_id AND s.status = 'active'
ORDER BY u.created_at DESC;

-- View for chat list with last message
CREATE VIEW chat_list AS
SELECT 
    c.id,
    c.name,
    c.type,
    c.avatar_url,
    c.member_count,
    c.last_message_at,
    c.created_at,
    m.content AS last_message_content,
    m.message_type AS last_message_type,
    u.username AS last_sender_username
FROM chats c
LEFT JOIN LATERAL (
    SELECT content, message_type, sender_id
    FROM messages
    WHERE chat_id = c.id AND is_deleted = FALSE
    ORDER BY created_at DESC
    LIMIT 1
) m ON TRUE
LEFT JOIN users u ON m.sender_id = u.id
ORDER BY c.last_message_at DESC;

-- =============================================
-- Initial Data (Optional)
-- =============================================

-- Insert system user for automated messages
INSERT INTO users (id, phone, email, username, password_hash, bio)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '+0000000000',
    'system@messenger.local',
    'system',
    '$2a$12$invalid_hash_for_system_user',
    'System automated messages'
) ON CONFLICT DO NOTHING;

-- =============================================
-- Vacuum and Analyze
-- =============================================

VACUUM ANALYZE;

-- =============================================
-- Comments for Documentation
-- =============================================

COMMENT ON TABLE users IS 'User accounts with authentication and profile data';
COMMENT ON TABLE chats IS 'Chat rooms (DM and group chats)';
COMMENT ON TABLE messages IS 'All messages sent in chats';
COMMENT ON TABLE channels IS 'Broadcast channels (one-to-many)';
COMMENT ON TABLE subscriptions IS 'Premium subscription records';
COMMENT ON TABLE payment_logs IS 'Payment history (MVP uses stub payments)';
COMMENT ON COLUMN payment_logs.payment_method IS 'MVP default: stub (no real payment processing)';
COMMENT ON COLUMN payment_logs.notes IS 'MVP: Stores "Stub payment - no real charge"';

-- Database setup complete
SELECT 'Database schema created successfully!' AS status;
