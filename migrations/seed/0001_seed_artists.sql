-- Seed initial artists data
USE album_store;

-- Insert default artist (for testing)
INSERT INTO artists (artist_id, name, created_at, updated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'Default Artist', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000002', 'Unknown Artist', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000003', 'Test Artist', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
