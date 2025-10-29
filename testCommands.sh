# Create User
grpcurl -plaintext -d '{\"user_id\": \"user123\", \"client_id\": \"client123\", \"account_type\": \"free\"}' localhost:50051 logsentinel.UserService/CreateUser

# Get User
grpcurl -plaintext -d '{\"user_id\": \"user123\"}' localhost:50051 logsentinel.UserService/GetUser

# Update User Account Type
grpcurl -plaintext -d '{\"user_id\": \"user123\", \"account_type\": \"pro\"}' localhost:50051 logsentinel.UserService/UpdateUserAccountType

# Delete User
grpcurl -plaintext -d '{\"user_id\": \"user123\"}' localhost:50051 logsentinel.UserService/DeleteUser

# Check User Quota
grpcurl -plaintext -d '{\"user_id\": \"user123\"}' localhost:50051 logsentinel.UserService/CheckUserQuota

# Create Project
grpcurl -plaintext -d '{\"name\": \"test-project\", \"user_id\": \"user123\", \"api_key\": \"test-api-key\"}' localhost:50051 logsentinel.ProjectService/CreateProject

# Get Project
grpcurl -plaintext -d '{\"project_id\": \"project-uuid\"}' localhost:50051 logsentinel.ProjectService/GetProject

# Delete Project
grpcurl -plaintext -d '{\"project_id\": \"project-uuid\"}' localhost:50051 logsentinel.ProjectService/DeleteProject

# Create Log
grpcurl -plaintext -d '{\"project_id\": \"project-uuid\", \"api_key\": \"api-key\", \"client_id\": \"client-id\", \"message\": \"Test log message\", \"category\": \"info\"}' localhost:50051 logsentinel.LogService/SendLog

# Stream Logs
grpcurl -plaintext -d '{\"project_id\": \"project-uuid\"}' localhost:50051 logsentinel.LogService/StreamLogs

# Test Stream
grpcurl -plaintext -d "{}" localhost:50051 logsentinel.LogService/Test

# Test Stream
grpcurl -plaintext localhost:50051 list

# Describe Service
grpcurl -plaintext localhost:50051 describe logsentinel.UserService