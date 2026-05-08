@{OPTIMIZER_VERSION=1}
SELECT SingerId, AlbumId, TrackId
FROM Songs
WHERE REGEXP_CONTAINS(SongName, "^A.*");

@{OPTIMIZER_VERSION=2}
SELECT SingerId, AlbumId, TrackId
FROM Songs
WHERE REGEXP_CONTAINS(SongName, "^A.*");

@{OPTIMIZER_VERSION=2}
SELECT s.SongGenre
FROM Songs AS s
ORDER BY SongGenre;

@{OPTIMIZER_VERSION=3}
SELECT s.SongGenre
FROM Songs AS s
ORDER BY SongGenre;

@{OPTIMIZER_VERSION=2}
SELECT
  t.ConcertDate,
  (
    SELECT COUNT(*)
    FROM UNNEST(t.TicketPrices) AS p
    WHERE p > 10
  ) AS expensive_tickets,
  u.VenueName
FROM Concerts AS t
JOIN Venues AS u ON t.VenueId = u.VenueId
ORDER BY expensive_tickets
LIMIT 2;

@{OPTIMIZER_VERSION=3}
SELECT
  t.ConcertDate,
  (
    SELECT COUNT(*)
    FROM UNNEST(t.TicketPrices) AS p
    WHERE p > 10
  ) AS expensive_tickets,
  u.VenueName
FROM Concerts AS t
JOIN Venues AS u ON t.VenueId = u.VenueId
ORDER BY expensive_tickets
LIMIT 2;

@{OPTIMIZER_VERSION=4}
SELECT AlbumTitle
FROM Songs
JOIN Albums ON Albums.AlbumId = Songs.AlbumId;

@{OPTIMIZER_VERSION=5}
SELECT AlbumTitle
FROM Songs
JOIN Albums ON Albums.AlbumId = Songs.AlbumId;

@{OPTIMIZER_VERSION=5}
SELECT a.SingerId, a.AlbumTitle, s.SongName
FROM Albums AS a
FULL OUTER JOIN Songs AS s
  ON a.SingerId = s.SingerId AND a.AlbumId = s.AlbumId
WHERE a.ReleaseDate >= DATE "2020-01-01" OR s.Duration > 180
LIMIT 10;

@{OPTIMIZER_VERSION=6}
SELECT a.SingerId, a.AlbumTitle, s.SongName
FROM Albums AS a
FULL OUTER JOIN Songs AS s
  ON a.SingerId = s.SingerId AND a.AlbumId = s.AlbumId
WHERE a.ReleaseDate >= DATE "2020-01-01" OR s.Duration > 180
LIMIT 10;

@{OPTIMIZER_VERSION=6}
SELECT s.SingerId
FROM Singers AS s
WHERE s.FirstName = "Alice" OR s.LastName = "Smith";

@{OPTIMIZER_VERSION=7}
SELECT s.SingerId
FROM Singers AS s
WHERE s.FirstName = "Alice" OR s.LastName = "Smith";

@{OPTIMIZER_VERSION=7, USE_UNENFORCED_FOREIGN_KEY=TRUE}
SELECT o.CustomerId
FROM FKOrders AS o
JOIN FKCustomers AS c ON c.CustomerId = o.CustomerId;

@{OPTIMIZER_VERSION=8, USE_UNENFORCED_FOREIGN_KEY=TRUE}
SELECT o.CustomerId
FROM FKOrders AS o
JOIN FKCustomers AS c ON c.CustomerId = o.CustomerId;
