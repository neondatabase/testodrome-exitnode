-- Create some indexes
CREATE INDEX queries_created_at_idx ON queries (created_at);
CREATE INDEX queries_region_id_created_at_idx ON queries (region_id, created_at);
CREATE INDEX queries_is_finished_is_failed_driver_exitnode_created_at_idx ON queries (is_finished, is_failed, driver, exitnode, created_at);
CREATE INDEX queries_project_id_created_at_idx ON queries (project_id, created_at);

INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (1, now(), now(), 'neon.tech', 'aws-us-east-1', 't');
INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (2, now(), now(), 'neon.tech', 'aws-us-east-2', 't');
INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (3, now(), now(), 'neon.tech', 'aws-us-west-2', 't');
INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (4, now(), now(), 'neon.tech', 'aws-eu-central-1', 't');
INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (5, now(), now(), 'neon.tech', 'aws-ap-southeast-1', 't');


INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (6, now(), now(), 'stage.neon.tech', 'aws-eu-west-1', 't');
INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (7, now(), now(), 'stage.neon.tech', 'aws-us-east-2', 't');
INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (8, now(), now(), 'neon.tech', 'aws-il-central-1', 't');


-- create new project every 10 minutes (in each region)
INSERT INTO global_rules("enabled", priority, "desc") VALUES (true, 1, '{"act": "create_project", "args": {"Interval": "10m"}}'::jsonb);

-- delete projects if there are > 5 (in each region)
INSERT INTO global_rules("enabled", priority, "desc") VALUES (true, 2, '{"act": "delete_project", "args": {"ProjectsN": 3, "SkipFailedQueries": {"Enabled": true, "QueriesN": 3}, "Matrix": ["projects.region_id", "projects.pg_version", "projects.provisioner", "projects.suspend_timeout_seconds"]}}'::jsonb);

-- query a random project
INSERT INTO global_rules("enabled", priority, "desc") VALUES (true, 3, '{"act": "query_project", "args": {"Scenario": "activityV1"}}'::jsonb);

-- start a forever connection
INSERT INTO global_rules("enabled", priority, "desc") VALUES (true, 4, '{"act": "query_project", "args": {"Scenario": "alwaysOn", "ConcurrencyLimit": 2}}'::jsonb);

-- start an inactive connection
INSERT INTO global_rules("enabled", priority, "desc") VALUES (true, 5, '{"act": "query_project", "args": {"Scenario": "awaitShutdown", "ConcurrencyLimit": 2}}'::jsonb);
