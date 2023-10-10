SELECT books.book_id, name, title, year
FROM books
INNER JOIN book_author
  ON books.book_id = book_author.book_id
INNER JOIN people
  ON book_author.author_id = people.person_id
WHERE people.name LIKE "Peter%Gentry"
  OR people.name LIKE "Stephen%Wellum"
  OR people.name LIKE "%Karen%Jobes%"
  OR people.name LIKE "%Gathercole%"
UNION
SELECT books.book_id, name, title, year
FROM books
INNER JOIN book_editor
  ON books.book_id = book_editor.book_id
INNER JOIN people
  ON book_editor.editor_id = people.person_id
WHERE people.name LIKE "Peter%Gentry"
  OR people.name LIKE "Stephen%Wellum"
  OR people.name LIKE "%Karen%Jobes%"
  OR people.name LIKE "%Gathercole%";
