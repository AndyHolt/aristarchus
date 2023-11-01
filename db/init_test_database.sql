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

INSERT INTO people (name)
VALUES
  ("R. K. Harrison"),
  ("Anselm"),
  ("Peter J. Gentry"),
  ("Stephen J. Wellum"),
  ("Herman Bavinck"),
  ("Robert J. Matz"),
  ("A. Chadwick Thornhill"),
  ("Thomas Williams"),
  ("N. Gray Sutanto"),
  ("James Eglinton"),
  ("Cory C. Brock");

INSERT INTO publishers (name)
VALUES
  ("IVP"),
  ("Hackett"),
  ("Crossway");

INSERT INTO series (series_name)
VALUES
  ("Spectrum Multiview Books");

INSERT INTO books (title, subtitle, year, edition, publisher_id, isbn,
series_id, status, purchased_date)
VALUES
  ("Introduction to the Old Testament", NULL, 1969,  NULL, 1, "0-85111-723-6", NULL, "Owned", "May 2023"),
  ("Divine Impassibility", "Four Views of God's Emotions and Suffering", 2019,  NULL, 1, "978-0-8308-5253-6", 1, "Owned", "October 2019"),
  ("Basic Writings", NULL, 2007, NULL, 2, "978-0-87220-895-7", NULL, "Owned", "October 2015"),
  ("How to Read and Understand the Biblical Prophets", NULL, 2017, NULL, 3, "978-1-4335-5403-8", NULL, "Owned", "July 2021"),
  ("Kingdom through Covenant", "A Biblical-Theological Understanding of the Covenants", 2018, 2, 3, "978-1-4335-5307-3", NULL, "Owned", "January 2022"),
  ("Christianity and Science", NULL, 2023, NULL, 3, "978-1-4335-7920-2", NULL, "Want", NULL);

INSERT INTO book_author (book_id, author_id)
VALUES
  (1, 1),
  (3, 2),
  (4, 3),
  (5, 3),
  (5, 4),
  (6, 5);

INSERT INTO book_editor (book_id, editor_id)
VALUES
  (2, 6),
  (2, 7),
  (3, 8),
  (6, 9),
  (6, 10),
  (6, 11);

.quit
