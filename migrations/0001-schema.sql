CREATE TABLE luser (
  luserId INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT UNIQUE NOT NULL,
  passwordHash TEXT NOT NULL,
  passwordSalt TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  avatarUrl TEXT NOT NULL -- pass gravatar url as default using https://gravatar.com/avatar/$(sha256 email)
);

CREATE TABLE community (
  communityId INTEGER PRIMARY KEY AUTOINCREMENT,
  communityName TEXT NOT NULL,
  createdByLuserId INTEGER REFERENCES luser(luserId),
  createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  default_ INTEGER DEFAULT 0
);

CREATE TABLE community_members (
  communityId INTEGER,
  memberId INTEGER,
  FOREIGN KEY (communityId) REFERENCES community(communityId)
  FOREIGN KEY (memberId) REFERENCES luser(luserId)
);

CREATE TABLE list (
  listId INTEGER PRIMARY KEY AUTOINCREMENT,
  title VARCHAR(255),
  description TEXT,
  creatorLuserId INTEGER REFERENCES user(luserId),
  updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  communityId INTEGER REFERENCES community(communityId)
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
  listId INTEGER,
  createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  name TEXT,
  FOREIGN KEY (listId) REFERENCES list(listId) ON DELETE SET NULL
);

CREATE TABLE list_group_items (
  itemId INTEGER PRIMARY KEY AUTOINCREMENT,
  groupId INTEGER,
  description TEXT,
  quantity INTEGER,
  order_ INTEGER,
  checked INTEGER DEFAULT 0,
  FOREIGN KEY (groupId) REFERENCES list_groups(groupId) ON DELETE CASCADE
);


CREATE TABLE luser_session (
 sessionId TEXT PRIMARY KEY,
 luserId INTEGER REFERENCES luser(luserId),
 lastUsed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE migrations (
  filename TEXT PRIMARY KEY,
  appliedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
