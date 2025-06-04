CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    threshold INTEGER NOT NULL
);

CREATE TABLE participants (
    id SERIAL PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    organization_id INTEGER NOT NULL,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);