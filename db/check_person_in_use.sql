SELECT COUNT(*)
FROM (
SELECT book_id
FROM book_author
WHERE author_id = 1
UNION
SELECT book_id
FROM book_editor
WHERE editor_id = 1
);
