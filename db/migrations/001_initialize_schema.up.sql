CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    threshold INTEGER NOT NULL
);

CREATE TABLE parties (
    id SERIAL PRIMARY KEY,
    address VARCHAR(255) NOT NULL
);

CREATE TABLE organization_parties (
    organization_id INTEGER NOT NULL,
    party_id       INTEGER NOT NULL,
    PRIMARY KEY (organization_id, party_id),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (party_id)       REFERENCES parties(id)       ON DELETE CASCADE
);