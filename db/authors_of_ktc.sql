SELECT people.name
FROM people
INNER JOIN book_author
  ON book_author.author_id = people.person_id
INNER JOIN books
  ON book_author.book_id = books.book_id
WHERE books.title = "Kingdom through Covenant";
