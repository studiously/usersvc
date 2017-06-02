-- +migrate Up
CREATE TYPE USER_ROLE AS ENUM ('student', 'teacher');

CREATE TABLE classes (
  id           UUID    NOT NULL PRIMARY KEY,
  name         TEXT    NOT NULL,
  current_unit UUID,
  active       BOOLEAN NOT NULL  DEFAULT TRUE
);


CREATE TABLE members (
  user_id  UUID      NOT NULL,
  class_id UUID      NOT NULL,
  role     USER_ROLE NOT NULL  DEFAULT 'student' :: USER_ROLE,
  owner    BOOLEAN   NOT NULL  DEFAULT FALSE,
  PRIMARY KEY(user_id, class_id),
  FOREIGN KEY ("class_id") REFERENCES classes ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX members_class_id_idx
  ON members USING BTREE (class_id);
CREATE UNIQUE INDEX members_user_id_class_id_idx
  ON members USING BTREE (user_id, class_id);
CREATE INDEX members_user_id_idx
  ON members USING BTREE (user_id);

-- +migrate Down
DROP TABLE classes;
DROP TABLE members;
DROP TYPE USER_ROLE;