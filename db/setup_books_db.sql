DROP TABLE IF EXISTS people;
CREATE TABLE people (
       person_id INTEGER PRIMARY KEY,
       name TEXT
);

DROP TABLE IF EXISTS publishers;
CREATE TABLE publishers (
       publisher_id INTEGER PRIMARY KEY,
       name TEXT
);

DROP TABLE IF EXISTS series;
CREATE TABLE series (
       series_id INTEGER PRIMARY KEY,
       series_name TEXT
);

DROP TABLE IF EXISTS books;
CREATE TABLE books (
       book_id INTEGER PRIMARY KEY,
       title TEXT NOT NULL,
       subtitle TEXT,
       year INTEGER,
       edition INTEGER,
       publisher_id INTEGER,
       isbn TEXT,
       series_id INTEGER,
       status TEXT NOT NULL,
       purchased_date TEXT,
       FOREIGN KEY (publisher_id)
         REFERENCES publishers (publisher_id)
           ON DELETE RESTRICT
           ON UPDATE CASCADE,
       FOREIGN KEY (series_id)
         REFERENCES series (series_id)
           ON DELETE RESTRICT
           ON UPDATE CASCADE
);

DROP TABLE IF EXISTS book_author;
CREATE TABLE book_author (
       book_id INTEGER,
       author_id INTEGER,
       PRIMARY KEY (book_id, author_id),
       FOREIGN KEY (book_id)
         REFERENCES books (book_id)
           ON DELETE CASCADE
           ON UPDATE CASCADE,
       FOREIGN KEY (author_id)
         REFERENCES people (person_id)
           ON DELETE RESTRICT
           ON UPDATE CASCADE
);

DROP TABLE IF EXISTS book_editor;
CREATE TABLE book_editor (
       book_id INTEGER,
       editor_id INTEGER,
       PRIMARY KEY (book_id, editor_id),
       FOREIGN KEY (book_id)
         REFERENCES books (book_id)
           ON DELETE CASCADE
           ON UPDATE CASCADE,
       FOREIGN KEY (editor_id)
         REFERENCES people (person_id)
           ON DELETE RESTRICT
           ON UPDATE CASCADE
);
