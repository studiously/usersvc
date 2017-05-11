-- +migrate Up

CREATE TABLE users (
  id     UUID                   NOT NULL PRIMARY KEY,
  name   TEXT                   NOT NULL,
  email  CHARACTER VARYING(320) NOT NULL,
  active BOOL                   NOT NULL DEFAULT TRUE,
  UNIQUE (email)
);

CREATE TABLE local_identities (
  user_id  UUID NOT NULL PRIMARY KEY,
  FOREIGN KEY ("user_id") REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
  password TEXT NOT NULL
);

-- +migrate Down

DROP TABLE local_identities;
DROP TABLE users;