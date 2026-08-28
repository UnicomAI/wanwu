-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- Create databases for all services
CREATE DATABASE IF NOT EXISTS kweaver CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
-- CREATE DATABASE IF NOT EXISTS adp CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

-- Grant permissions
GRANT ALL PRIVILEGES ON kweaver.* TO 'root'@'%';
-- GRANT ALL PRIVILEGES ON adp.* TO 'root'@'%';

FLUSH PRIVILEGES;


