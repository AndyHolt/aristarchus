SELECT name, title, subtitle, year, status
FROM books
INNER JOIN book_author
  ON book_author.book_id = books.book_id
INNER JOIN people
  ON book_author.author_id = people.person_id
WHERE name == "Peter J. Gentry";
