# DynamoDB Table Design - Composite Key Explanation

## The Problem You Identified

You asked: "If session_id is the primary key and it's unique, how can I store multiple messages with the same session_id?"

**Answer**: Use a **Composite Primary Key** (Partition Key + Sort Key)

## How Composite Keys Work

### Simple Primary Key (WRONG for messages)
```
Primary Key: session_id only

session_id | message
-----------|---------
session-1  | "Hello"
session-1  | "How are you?"  ❌ ERROR! Duplicate key
```
**Result**: Can only store 1 message per session

### Composite Primary Key (CORRECT)
```
Partition Key: session_id
Sort Key: timestamp

session_id | timestamp        | message
-----------|------------------|------------------
session-1  | 1704067200000000 | "Hello"           ✅
session-1  | 1704067201000000 | "How are you?"    ✅
session-1  | 1704067202000000 | "I'm fine"        ✅
session-2  | 1704067200000000 | "Different user"  ✅
```
**Result**: Multiple messages per session, each with unique timestamp

## Uniqueness Rule

In DynamoDB, the **combination** of partition key + sort key must be unique:

```
✅ ALLOWED:
- (session-1, timestamp-100)
- (session-1, timestamp-101)  // Same session, different timestamp
- (session-2, timestamp-100)  // Different session, same timestamp

❌ REJECTED:
- (session-1, timestamp-100)
- (session-1, timestamp-100)  // Exact duplicate - will overwrite!
```

## Why We Use UnixNano() Timestamp

```go
Timestamp: time.Now().UnixNano()
```

**UnixNano() returns nanoseconds**: `1704067200123456789`
- Precision: 1 nanosecond (0.000000001 seconds)
- Collision chance: Nearly impossible (would need 2 messages in same nanosecond)

**Alternatives and their problems:**
- `time.Now().Unix()` - seconds - ❌ Very high collision chance
- `time.Now().UnixMilli()` - milliseconds - ⚠️ Possible collisions under load
- `time.Now().UnixNano()` - nanoseconds - ✅ Virtually collision-free

## Table Schema Details

### Messages Table
```
Table Name: messages

Primary Key (Composite):
  Partition Key: session_id (String)
  Sort Key: timestamp (Number) - nanoseconds since Unix epoch

Attributes:
  - id (String) - UUID for reference/debugging
  - session_id (String) - PARTITION KEY
  - timestamp (Number) - SORT KEY (nanoseconds)
  - message (String) - The actual message content
  - interaction (String) - Role: "user", "model", or "function"
  - business_phone_number (String)
  - customer_phone_number (String)
  - created_at (String) - ISO 8601 timestamp for humans

Billing Mode: PAY_PER_REQUEST (on-demand)
```

### Sessions Table (unchanged)
```
Table Name: sessions

Primary Key (Simple):
  Partition Key: id (String) - session ID

Attributes:
  - id (String) - PRIMARY KEY (generated from business + customer phone)
  - business_phone_number (String)
  - customer_phone_number (String)
  - created_at (String)
  - updated_at (String)
  - session_expiry_at (String)

Billing Mode: PAY_PER_REQUEST
```

### Customers Table (unchanged)
```
Table Name: customers

Primary Key (Simple):
  Partition Key: business_phone_number (String)

Attributes:
  - business_phone_number (String) - PRIMARY KEY
  - created_at (String)
  - updated_at (String)
  - ai_prompt (String)
  - products_info (String)

Billing Mode: PAY_PER_REQUEST
```

## Query Performance

### Before (Scan with Filter)
```go
Scan entire table → Filter by session_id → Return results
Time: 300ms - 5000ms (depending on table size)
Cost: Reads entire table
```

### After (Query with Composite Key)
```go
Query directly by session_id → Results already sorted by timestamp
Time: 10ms - 30ms (constant, regardless of table size)
Cost: Only reads messages for that session
```

**Performance improvement**: 10x - 100x faster

## Example Data in DynamoDB

```json
{
  "session_id": "1234567890-573001234567",
  "timestamp": 1704067200123456789,
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Hello, I need help",
  "interaction": "user",
  "business_phone_number": "1234567890",
  "customer_phone_number": "573001234567",
  "created_at": "2024-01-01T00:00:00.123Z"
}

{
  "session_id": "1234567890-573001234567",
  "timestamp": 1704067202456789012,
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "message": "How can I help you today?",
  "interaction": "model",
  "business_phone_number": "1234567890",
  "customer_phone_number": "573001234567",
  "created_at": "2024-01-01T00:00:02.456Z"
}
```

Notice:
- Same `session_id` ✅
- Different `timestamp` ✅
- Unique combination allows both to exist

## Migration Steps

### Option 1: Fresh Start (No existing data)
```bash
# Delete old table if it exists
aws dynamodb delete-table --table-name messages

# Create new table with composite key
./scripts/create-all-tables.sh
```

### Option 2: Migrate Existing Data
```bash
# 1. Create new table with different name
aws dynamodb create-table --table-name messages_v2 ...

# 2. Scan old table and copy data
aws dynamodb scan --table-name messages > old_data.json

# 3. Transform and insert into new table
# (write a migration script)

# 4. Switch application to use new table
# 5. Delete old table
```

### Option 3: Use GSI (Keep old structure)
If you can't change the primary key:
```bash
# Add Global Secondary Index
aws dynamodb update-table \
    --table-name messages \
    --attribute-definitions \
        AttributeName=session_id,AttributeType=S \
        AttributeName=timestamp,AttributeType=N \
    --global-secondary-index-updates \
        '[{
            "Create": {
                "IndexName": "session-timestamp-index",
                "KeySchema": [
                    {"AttributeName": "session_id", "KeyType": "HASH"},
                    {"AttributeName": "timestamp", "KeyType": "RANGE"}
                ],
                "Projection": {"ProjectionType": "ALL"},
                "ProvisionedThroughput": {
                    "ReadCapacityUnits": 5,
                    "WriteCapacityUnits": 5
                }
            }
        }]'
```

Then update query code to use the index:
```go
IndexName: aws.String("session-timestamp-index"),
```

## Summary

✅ **Composite Key = Partition Key + Sort Key**
✅ **Uniqueness = Combination must be unique**
✅ **session_id + timestamp = Multiple messages per session**
✅ **UnixNano() = Collision-free timestamps**
✅ **Query instead of Scan = 10-100x faster**
