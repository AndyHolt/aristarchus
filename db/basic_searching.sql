SELECT books.book_id, title, year
FROM books
INNER JOIN book_author
  ON books.book_id = book_author.book_id
INNER JOIN people
  ON book_author.author_id = people.person_id
WHERE people.name LIKE "%Gentry%" AND title LIKE %%;
