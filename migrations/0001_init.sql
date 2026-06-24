-- Party reference-data schema. Owned exclusively by the party microservice;
-- no other service reads or writes these tables.
CREATE SCHEMA IF NOT EXISTS party;

CREATE TABLE IF NOT EXISTS party.parties (
    party_id        TEXT PRIMARY KEY,
    party_type      TEXT    NOT NULL,
    name            TEXT    NOT NULL,
    full_legal_name TEXT    NOT NULL DEFAULT '',
    email_address   TEXT    NOT NULL DEFAULT '',
    phone           TEXT    NOT NULL DEFAULT '',

    -- simple single postal address
    address_line    TEXT    NOT NULL DEFAULT '',
    post_code       TEXT    NOT NULL DEFAULT '',
    country         TEXT    NOT NULL DEFAULT '',

    -- exactly one row is the primary PSU party served by GET /party
    primary_party   BOOLEAN NOT NULL DEFAULT false
);

-- Demo PSU parties. The primary party (PSU-001) is "Mr Kelvin C Smith",
-- matching the accounts service demo PSU. Re-running is a no-op.
INSERT INTO party.parties
    (party_id, party_type, name, full_legal_name, email_address, phone,
     address_line, post_code, country, primary_party)
VALUES
    ('PSU-001', 'Individual', 'Kelvin Smith', 'Mr Kelvin C Smith',
     'kelvin.smith@example.com', '+44 7700 900123',
     '1 Bank Street', 'EC2R 8AH', 'GB', true),
    ('PARTY-002', 'Individual', 'Amy Jones', 'Ms Amy R Jones',
     'amy.jones@example.com', '+44 7700 900456',
     '27 High Holborn', 'WC1V 6AA', 'GB', false)
ON CONFLICT (party_id) DO NOTHING;
