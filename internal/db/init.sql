CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE account_type AS ENUM ('free', 'pro');

-- users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) UNIQUE NOT NULL,
    account_type account_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
]
CREATE TYPE log_category AS ENUM ('error', 'warning', 'info', 'event', 'system');

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Logs table
CREATE TABLE IF NOT EXISTS logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    category log_category NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS logs_created_at_idx ON logs(created_at);
CREATE INDEX IF NOT EXISTS logs_project_id_idx ON logs(project_id);


-- CREATE OR REPLACE FUNCTION delete_old_logs() RETURNS void AS $$
-- BEGIN
--     DELETE FROM logs l
--     USING projects p, users u
--     WHERE l.project_id = p.id 
--     AND p.user_id = u.id 
--     AND u.account_type = 'free'
--     AND l.created_at < NOW() - INTERVAL '7 days';
-- END;

-- -- Create a scheduled job to run delete_old_logs() daily
-- CREATE EXTENSION IF NOT EXISTS pg_cron;
-- SELECT cron.schedule('0 0 * * *', 'SELECT delete_old_logs()');
-- -- Create extensions if needed
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- CREATE TYPE account_type AS ENUM ('free', 'pro');

-- CREATE TABLE IF NOT EXISTS users (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     user_id VARCHAR(255) UNIQUE NOT NULL,
--     account_type account_type NOT NULL,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );

-- CREATE TYPE log_category AS ENUM ('error', 'warning', 'info', 'event', 'system');

-- -- Teams table
-- CREATE TABLE IF NOT EXISTS teams (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     -- id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

--     name VARCHAR(255) NOT NULL,
--     owner_id VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

-- -- Projects table
-- CREATE TABLE IF NOT EXISTS projects (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     name VARCHAR(255) NOT NULL,
--     team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
--     api_key VARCHAR(255) UNIQUE NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

-- -- Logs table
-- CREATE TABLE IF NOT EXISTS logs (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
--     category log_category NOT NULL,
--     message TEXT NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

-- CREATE INDEX IF NOT EXISTS logs_created_at_idx ON logs(created_at);
-- CREATE INDEX IF NOT EXISTS logs_project_id_idx ON logs(project_id);