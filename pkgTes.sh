# Test BatchSendLogs
cat << EOF | grpcurl -plaintext -d @ localhost:50051 logsentinel.LogService/BatchSendLogs
{"project_id": "006a99c0-f1ea-4ca4-bc75-a5572393925a", "api_key": "test-api-key", "client_id": "client123", "message": "First batch log message", "category": "info"}
{"project_id": "006a99c0-f1ea-4ca4-bc75-a5572393925a", "api_key": "test-api-key", "client_id": "client123", "message": "Second batch log message", "category": "error"}
{"project_id": "006a99c0-f1ea-4ca4-bc75-a5572393925a", "api_key": "test-api-key", "client_id": "client123", "message": "Third batch log message", "category": "warning"}
EOF