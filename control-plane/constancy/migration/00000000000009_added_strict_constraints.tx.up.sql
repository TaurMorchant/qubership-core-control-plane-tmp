ALTER TABLE routes ALTER COLUMN deployment_version SET DEFAULT 'v1';
ALTER TABLE routes ALTER COLUMN deployment_version SET NOT NULL;
ALTER TABLE routes ALTER COLUMN initialDeploymentVersion SET DEFAULT 'v1';
ALTER TABLE routes ALTER COLUMN initialDeploymentVersion SET NOT NULL;
ALTER TABLE routes ALTER COLUMN version SET DEFAULT 0;
ALTER TABLE routes ALTER COLUMN version SET NOT NULL;