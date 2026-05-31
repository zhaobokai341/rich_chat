# Database Migration Guide

## Problem
The account lockout feature was not working correctly due to timezone mismatches between:
- System time (CST - China Standard Time, UTC+8)
- PostgreSQL timezone (EDT - Eastern Daylight Time, UTC-4)
- Go's `time.Now()` using system local time

This caused accounts to remain locked even after the lock expiration time had passed.

## Solution
Convert all timestamp columns from `TIMESTAMP` to `TIMESTAMPTZ` (timestamp with timezone) and add performance indexes.

## Benefits
1. **Fixes timezone issues**: TIMESTAMPTZ automatically handles timezone conversions
2. **Better performance**: Added indexes for frequently queried columns
3. **No code changes needed**: Database handles timezone conversion automatically
4. **Data integrity**: Existing data is preserved and converted correctly

## Migration Steps

### For New Databases
Simply run the updated `setup.sql`:
```bash
psql -U postgres -f script/setup.sql
```

### For Existing Databases
Run the migration script:
```bash
psql -U postgres -d rich_chat -f script/migrate_to_timestamptz.sql
```

This will:
1. Convert all TIMESTAMP columns to TIMESTAMPTZ
2. Interpret existing timestamps as UTC
3. Add performance indexes
4. Verify the changes

## Indexes Added

### Users Table
- `idx_users_username`: Speeds up username lookups during login
- `idx_users_lock_until`: Partial index for lock status checks (only non-NULL values)

### IP Block List Table
- `idx_ip_block_list_ip`: Speeds up IP blocking checks
- `idx_ip_block_list_block_until`: Speeds up expired block cleanup

## Verification
After migration, verify the changes:
```sql
-- Check column types
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'users' 
AND column_name IN ('registered_time', 'last_login', 'lock_until');

-- Check indexes
SELECT indexname, tablename 
FROM pg_indexes 
WHERE schemaname = 'public' 
ORDER BY tablename, indexname;
```

## Rollback (if needed)
If you need to rollback:
```sql
ALTER TABLE users 
    ALTER COLUMN registered_time TYPE TIMESTAMP,
    ALTER COLUMN last_login TYPE TIMESTAMP,
    ALTER COLUMN lock_until TYPE TIMESTAMP;

ALTER TABLE ip_block_list
    ALTER COLUMN block_until TYPE TIMESTAMP,
    ALTER COLUMN created_time TYPE TIMESTAMP;

DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_lock_until;
DROP INDEX IF EXISTS idx_ip_block_list_ip;
DROP INDEX IF EXISTS idx_ip_block_list_block_until;
```
