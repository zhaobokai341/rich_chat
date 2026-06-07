CREATE TABLE IF NOT EXISTS ip_block_list (
    ip VARCHAR(50) NOT NULL,
    block_until TIMESTAMPTZ NOT NULL,
    reason TEXT,
    created_time TIMESTAMPTZ DEFAULT NOW()
);

-- Add index for IP blocking lookups
CREATE INDEX IF NOT EXISTS idx_ip_block_list_ip ON ip_block_list(ip);
CREATE INDEX IF NOT EXISTS idx_ip_block_list_block_until ON ip_block_list(block_until);
