-- +migrate Up

CREATE TABLE users (
  id UUID NOT NULL PRIMARY KEY,
  name text NOT NULL,
  email character varying(320) NOT NULL
);

CREATE TABLE local_identities (
  user_id UUID NOT NULL PRIMARY KEY,
  FOREIGN KEY ("user_id") REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
  password text NOT NULL
);

-- +migrate Down

DROP TABLE local_identities;
DROP TABLE users;