INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (1, now(), now(), 'neon.tech', 'aws-us-east-1', 't');
INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (2, now(), now(), 'neon.tech', 'aws-us-east-2', 't');
INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (3, now(), now(), 'neon.tech', 'aws-us-west-2', 't');
INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (4, now(), now(), 'neon.tech', 'aws-eu-central-1', 't');
INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (5, now(), now(), 'neon.tech', 'aws-ap-southeast-1', 't');


INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (6, now(), now(), 'stage.neon.tech', 'aws-eu-west-1', 'f');
INSERT INTO regions(id, created_at, updated_at, "provider", database_region, supports_neon_vm) VALUES (7, now(), now(), 'stage.neon.tech', 'aws-us-east-2', 't');


-- create new project every 10 minutes (in each region)
INSERT INTO global_rules("enabled", priority, "desc") VALUES (true, 1, '{"act": "create_project", "args": {"Interval": "10m"}}'::jsonb);

-- delete projects if there are > 5 (in each region)
INSERT INTO global_rules("enabled", priority, "desc") VALUES (true, 2, '{"act": "delete_project", "args": {"ProjectsN": 5}}'::jsonb);

-- query a random project
INSERT INTO global_rules("enabled", priority, "desc") VALUES (true, 3, '{"act": "query_project", "args": {}}'::jsonb);
