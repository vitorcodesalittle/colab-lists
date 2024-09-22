CREATE TABLE luser (
  luserId INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL,
  passwordHash TEXT NOT NULL,
  passwordSalt TEXT NOT NULL,
  UNIQUE(username)
);
CREATE TABLE list (
    listId INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(255),
    description TEXT,
    creatorLuserId INT,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (creatorLuserId) REFERENCES user(luserId)
);

CREATE TABLE list_colaborators (
  listId INTEGER,
  luserId INTEGER,
  createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (listId, luserId),
  FOREIGN KEY (luserId) REFERENCES luser(luserId),
  FOREIGN KEY (listId) REFERENCES list(listId)
);
CREATE TABLE list_groups (
  groupId INTEGER PRIMARY KEY AUTOINCREMENT,
  listId INT,
  createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  name TEXT,
  FOREIGN KEY (listId) REFERENCES list(listId) ON DELETE SET NULL
);

CREATE TABLE list_group_items (
  itemId INTEGER PRIMARY KEY AUTOINCREMENT,
  groupId INT,
  description TEXT,
  quantity INT,
  order_ INT,
  FOREIGN KEY (groupId) REFERENCES list_groups(groupId) ON DELETE CASCADE
);


CREATE TABLE luser_session (
    sessionId TEXT PRIMARY KEY,
    luserId INTEGER,
    lastUsed TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    -- FOREIGN KEY (luserId) REFERENCES luser(luserId)
);
