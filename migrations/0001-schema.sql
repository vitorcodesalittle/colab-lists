CREATE TABLE luser (
  luserId INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT UNIQUE NOT NULL,
  passwordHash TEXT NOT NULL,
  passwordSalt TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  avatarUrl TEXT NOT NULL -- pass gravatar url as default using https://gravatar.com/avatar/$(sha256 email)
);

CREATE TABLE houses (
        houseId INTEGER PRIMARY KEY AUTOINCREMENT,
        houseName TEXT NOT NULL,
        createdByLuserId INTEGER,
        FOREIGN KEY (createdByLuserId) REFERENCES luser(luserId)
    );

    CREATE TABLE houser_members (
        houseId INTEGER,
        memberId INTEGER,
        FOREIGN KEY (houseId) REFERENCES houses(houseId)
        FOREIGN KEY (memberId) REFERENCES luser(luserId)
    );

    CREATE TABLE list (
        listId INTEGER PRIMARY KEY AUTOINCREMENT,
        title VARCHAR(255),
        description TEXT,
        creatorLuserId INTEGER REFERENCES user(luserId),
        updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        houseId INTEGER REFERENCES houses(houseId)
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
      FOREIGN KEY (groupId) REFERENCES list_groups(groupId) ON DELETE CASCADE
    );


    CREATE TABLE luser_session (
        sessionId TEXT PRIMARY KEY,
        luserId INTEGER,
    lastUsed TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    -- FOREIGN KEY (luserId) REFERENCES luser(luserId)
);

CREATE TABLE migrations (
    filename TEXT PRIMARY KEY,
    appliedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
