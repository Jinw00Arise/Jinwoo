-- +goose Up
-- Initial schema for Jinwoo MapleStory Server

CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    gm_level INTEGER DEFAULT 0,
    ban_reason TEXT,
    banned_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS characters (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES accounts(id),
    world_id SMALLINT NOT NULL,
    name VARCHAR(13) UNIQUE NOT NULL,
    gender SMALLINT DEFAULT 0,
    skin_color SMALLINT DEFAULT 0,
    face INTEGER DEFAULT 20000,
    hair INTEGER DEFAULT 30000,
    level SMALLINT DEFAULT 1,
    job SMALLINT DEFAULT 0,
    str SMALLINT DEFAULT 12,
    dex SMALLINT DEFAULT 5,
    "int" SMALLINT DEFAULT 4,
    luk SMALLINT DEFAULT 4,
    hp INTEGER DEFAULT 50,
    max_hp INTEGER DEFAULT 50,
    mp INTEGER DEFAULT 5,
    max_mp INTEGER DEFAULT 5,
    ap SMALLINT DEFAULT 0,
    sp SMALLINT DEFAULT 0,
    exp INTEGER DEFAULT 0,
    fame SMALLINT DEFAULT 0,
    map_id INTEGER DEFAULT 0,
    spawn_point SMALLINT DEFAULT 0,
    meso INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_characters_account_world ON characters(account_id, world_id);

CREATE TABLE IF NOT EXISTS inventories (
    id SERIAL PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES characters(id),
    type SMALLINT NOT NULL,
    slot SMALLINT NOT NULL,
    item_id INTEGER NOT NULL,
    quantity SMALLINT DEFAULT 1,
    expire_time TIMESTAMP,
    UNIQUE(character_id, type, slot)
);

CREATE INDEX idx_inventories_character ON inventories(character_id);

CREATE TABLE IF NOT EXISTS quest_records (
    id SERIAL PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES characters(id),
    quest_id INTEGER NOT NULL,
    state SMALLINT DEFAULT 0,
    progress VARCHAR(512),
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(character_id, quest_id)
);

CREATE INDEX idx_quest_records_character ON quest_records(character_id);

CREATE TABLE IF NOT EXISTS quest_record_exs (
    id SERIAL PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES characters(id),
    quest_id INTEGER NOT NULL,
    value VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(character_id, quest_id)
);

CREATE TABLE IF NOT EXISTS skills (
    id SERIAL PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES characters(id),
    skill_id INTEGER NOT NULL,
    skill_level INTEGER DEFAULT 0,
    master_level INTEGER DEFAULT 0,
    expire_time TIMESTAMP,
    UNIQUE(character_id, skill_id)
);

CREATE INDEX idx_skills_character ON skills(character_id);

CREATE TABLE IF NOT EXISTS skill_macros (
    id SERIAL PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES characters(id),
    position SMALLINT NOT NULL,
    name VARCHAR(255),
    shout BOOLEAN DEFAULT FALSE,
    skill1 INTEGER DEFAULT 0,
    skill2 INTEGER DEFAULT 0,
    skill3 INTEGER DEFAULT 0,
    UNIQUE(character_id, position)
);

CREATE TABLE IF NOT EXISTS key_bindings (
    id SERIAL PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES characters(id),
    key_index INTEGER NOT NULL,
    key_type SMALLINT DEFAULT 0,
    action INTEGER DEFAULT 0,
    UNIQUE(character_id, key_index)
);

CREATE TABLE IF NOT EXISTS quick_slots (
    id SERIAL PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES characters(id),
    slot_index SMALLINT NOT NULL,
    key_code INTEGER DEFAULT 0,
    UNIQUE(character_id, slot_index)
);

-- +goose Down
DROP TABLE IF EXISTS quick_slots;
DROP TABLE IF EXISTS key_bindings;
DROP TABLE IF EXISTS skill_macros;
DROP TABLE IF EXISTS skills;
DROP TABLE IF EXISTS quest_record_exs;
DROP TABLE IF EXISTS quest_records;
DROP TABLE IF EXISTS inventories;
DROP TABLE IF EXISTS characters;
DROP TABLE IF EXISTS accounts;

