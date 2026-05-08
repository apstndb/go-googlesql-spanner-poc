# spanner-hacks optimizer version feedback candidates

This note collects optimizer-version plan-shape diffs that are good candidates
for `spanner-hacks` `operators.md` follow-up material.

Generated on 2026-05-08 with:

- `github.com/apstndb/spannerplan v0.1.9`
- `github.com/apstndb/spanemuboost v0.4.0`
- `go1.26.2 darwin/arm64`

Official source checked with `dkcli get`:

- <https://docs.cloud.google.com/spanner/docs/query-optimizer/versions>
- `updateTime: 2026-05-03T18:17:15Z`
- Current default and latest optimizer version: `8`

Commands:

```sh
go run ./tools/spanner-query-plan-shape \
  --case docs \
  --optimizer-version-matrix \
  --output reference \
  --continue-on-error \
  > /private/tmp/spanner-docs-optimizer-reference.txt

go run ./tools/spanner-query-plan-shape \
  --case optimizer_gaps \
  --optimizer-version-matrix \
  --output reference \
  --continue-on-error \
  > /private/tmp/spanner-optimizer-gaps-reference.txt

go run ./tools/spanner-query-plan-shape \
  --case docs \
  --sql-file /private/tmp/regexp-v1.sql \
  --sql-file /private/tmp/regexp-v2.sql \
  --sql-file /private/tmp/like-v1.sql \
  --sql-file /private/tmp/like-v2.sql \
  --output reference \
  --continue-on-error \
  > /private/tmp/spanner-v1-v2-prefix-reference.txt
```

## Suggested scope

The upstream page should probably not include the whole matrix. The useful
additions are representative before/after pairs where a documented optimizer
version item is visible in the rendered plan:

- v3: distributed merge / sorted limit placement.
- v4: secondary index selection for a predicate on the leading indexed column.
- v5: cost-based join algorithm selection.
- v6: full outer join predicate and limit pushdown.
- v7: cost-based index union.
- v8: foreign-key handling with `USE_UNENFORCED_FOREIGN_KEY=TRUE`.

Version 4 was a weak candidate in the local synthetic database. The DBaaS
example below gives a compact before/after pair for the documented
secondary-index selection improvement, but a minimal Spanner Omni 2026.r1-beta
recheck with the same schema did not reproduce it. Treat the v4 example as
DBaaS evidence, not as an Omni-reproduced shape yet.

## v2: prefix extraction from `REGEXP_CONTAINS` and `LIKE`

Official item: version 2 improved `REGEXP_CONTAINS` and `LIKE` predicates under
certain circumstances. These two examples show v2 extracting a `STARTS_WITH`
range from a prefix pattern when scanning `SongsBySongName`.

```text
=== /private/tmp/regexp-v1.sql ===
@{OPTIMIZER_VERSION=1}
SELECT SingerId, AlbumId, TrackId
FROM Songs@{FORCE_INDEX=SongsBySongName}
WHERE REGEXP_CONTAINS(SongName,"^A.*");
+----+-------------------------------------------------------------------------------------+
| ID | Operator                                                                            |
+----+-------------------------------------------------------------------------------------+
| *0 | Distributed Union on SongsBySongName <Row>                                          |
|  1 | +- Local Distributed Union <Row>                                                    |
|  2 |    +- Serialize Result <Row>                                                        |
| *3 |       +- Filter Scan <Row> (seekable_key_size: 0)                                   |
|  4 |          +- Index Scan on SongsBySongName <Row> (Full scan, scan_method: Automatic) |
+----+-------------------------------------------------------------------------------------+
Predicates(identified by ID):
 0: Split Range: REGEXP_CONTAINS($SongName, '^A.*')
 3: Residual Condition: REGEXP_CONTAINS($SongName, '^A.*')


=== /private/tmp/regexp-v2.sql ===
@{OPTIMIZER_VERSION=2}
SELECT SingerId, AlbumId, TrackId
FROM Songs@{FORCE_INDEX=SongsBySongName}
WHERE REGEXP_CONTAINS(SongName,"^A.*");
+----+--------------------------------------------------------------------+
| ID | Operator                                                           |
+----+--------------------------------------------------------------------+
| *0 | Distributed Union on SongsBySongName <Row>                         |
|  1 | +- Local Distributed Union <Row>                                   |
|  2 |    +- Serialize Result <Row>                                       |
|  3 |       +- Filter Scan <Row> (seekable_key_size: 1)                  |
| *4 |          +- Index Scan on SongsBySongName <Row> (scan_method: Row) |
+----+--------------------------------------------------------------------+
Predicates(identified by ID):
 0: Split Range: STARTS_WITH($SongName, 'A')
 4: Seek Condition: STARTS_WITH($SongName, 'A')
```

```text
=== /private/tmp/like-v1.sql ===
@{OPTIMIZER_VERSION=1}
SELECT SingerId, AlbumId, TrackId
FROM Songs@{FORCE_INDEX=SongsBySongName}
WHERE SongName LIKE "A%z";
+----+-------------------------------------------------------------------------------------+
| ID | Operator                                                                            |
+----+-------------------------------------------------------------------------------------+
| *0 | Distributed Union on SongsBySongName <Row>                                          |
|  1 | +- Local Distributed Union <Row>                                                    |
|  2 |    +- Serialize Result <Row>                                                        |
| *3 |       +- Filter Scan <Row> (seekable_key_size: 0)                                   |
|  4 |          +- Index Scan on SongsBySongName <Row> (Full scan, scan_method: Automatic) |
+----+-------------------------------------------------------------------------------------+
Predicates(identified by ID):
 0: Split Range: ($SongName LIKE 'A%z')
 3: Residual Condition: ($SongName LIKE 'A%z')


=== /private/tmp/like-v2.sql ===
@{OPTIMIZER_VERSION=2}
SELECT SingerId, AlbumId, TrackId
FROM Songs@{FORCE_INDEX=SongsBySongName}
WHERE SongName LIKE "A%z";
+----+--------------------------------------------------------------------+
| ID | Operator                                                           |
+----+--------------------------------------------------------------------+
| *0 | Distributed Union on SongsBySongName <Row>                         |
|  1 | +- Local Distributed Union <Row>                                   |
|  2 |    +- Serialize Result <Row>                                       |
| *3 |       +- Filter Scan <Row> (seekable_key_size: 1)                  |
| *4 |          +- Index Scan on SongsBySongName <Row> (scan_method: Row) |
+----+--------------------------------------------------------------------+
Predicates(identified by ID):
 0: Split Range: (STARTS_WITH($SongName, 'A') AND ($SongName LIKE 'A%z'))
 3: Residual Condition: ($SongName LIKE 'A%z')
 4: Seek Condition: STARTS_WITH($SongName, 'A')
```

## v3: distributed merge union for sorted results

Official item: version 3 introduced distributed merge union behavior. In raw
plans, this appears as `Distributed Union` with
`preserve_subquery_order: true`, not as a separate `PlanNode.display_name`.

```text
=== optimizer-version/v2/unary/sort ===
@{OPTIMIZER_VERSION=2}
SELECT s.SongGenre FROM Songs AS s ORDER BY SongGenre
+----+---------------------------------------------------------------------------+
| ID | Operator                                                                  |
+----+---------------------------------------------------------------------------+
|  0 | Serialize Result <Row>                                                    |
|  1 | +- Sort <Row>                                                             |
|  2 |    +- Distributed Union on Songs <Row>                                    |
|  3 |       +- Local Distributed Union <Row>                                    |
|  4 |          +- Table Scan on Songs <Row> (Full scan, scan_method: Automatic) |
+----+---------------------------------------------------------------------------+


=== optimizer-version/v3/unary/sort ===
@{OPTIMIZER_VERSION=3}
SELECT s.SongGenre FROM Songs AS s ORDER BY SongGenre
+----+---------------------------------------------------------------------------+
| ID | Operator                                                                  |
+----+---------------------------------------------------------------------------+
|  0 | Distributed Union on Songs <Row> (preserve_subquery_order: true)          |
|  1 | +- Serialize Result <Row>                                                 |
|  2 |    +- Sort <Row>                                                          |
|  3 |       +- Local Distributed Union <Row>                                    |
|  4 |          +- Table Scan on Songs <Row> (Full scan, scan_method: Automatic) |
+----+---------------------------------------------------------------------------+
```

## v3: sorted LIMIT across Cross Apply

Official item: version 3 improved sorted `LIMIT` placement when `Cross Apply`
is introduced by joins.

```text
=== optimizer-version/v2/optimizer-gaps/v3/sorted-limit-cross-apply ===
@{OPTIMIZER_VERSION=2}
SELECT a2.*
FROM Albums@{FORCE_INDEX=_BASE_TABLE} AS a1
JOIN Albums@{FORCE_INDEX=_BASE_TABLE} AS a2 USING(SingerId)
ORDER BY a1.AlbumId
LIMIT 2
+-----+------------------------------------------------------------------------------------------+
| ID  | Operator                                                                                 |
+-----+------------------------------------------------------------------------------------------+
|   0 | Serialize Result <Row>                                                                   |
|   1 | +- Global Sort Limit <Row>                                                               |
|   2 |    +- Distributed Union on Singers <Row> (split_ranges_aligned)                          |
|   3 |       +- Local Sort Limit <Row>                                                          |
|   4 |          +- Local Distributed Union <Row>                                                |
|   5 |             +- Cross Apply <Row>                                                         |
|   6 |                +- [Input] Table Scan on Albums <Row> (Full scan, scan_method: Automatic) |
|   9 |                +- [Map] Local Distributed Union <Row>                                    |
|  10 |                   +- Filter Scan <Row> (seekable_key_size: 0)                            |
| *11 |                      +- Table Scan on Albums <Row> (scan_method: Row)                    |
+-----+------------------------------------------------------------------------------------------+
Predicates(identified by ID):
 11: Seek Condition: ($SingerId_1 = $SingerId)


=== optimizer-version/v3/optimizer-gaps/v3/sorted-limit-cross-apply ===
@{OPTIMIZER_VERSION=3}
SELECT a2.*
FROM Albums@{FORCE_INDEX=_BASE_TABLE} AS a1
JOIN Albums@{FORCE_INDEX=_BASE_TABLE} AS a2 USING(SingerId)
ORDER BY a1.AlbumId
LIMIT 2
+-----+---------------------------------------------------------------------------------------------+
| ID  | Operator                                                                                    |
+-----+---------------------------------------------------------------------------------------------+
|   0 | Global Limit <Row>                                                                          |
|   1 | +- Distributed Union on Singers <Row> (split_ranges_aligned, preserve_subquery_order: true) |
|   2 |    +- Serialize Result <Row>                                                                |
|   3 |       +- Local Limit <Row>                                                                  |
|   4 |          +- Cross Apply <Row>                                                               |
|   5 |             +- [Input] Local Sort Limit <Row>                                               |
|   6 |             |  +- Local Distributed Union <Row>                                             |
|   7 |             |     +- Table Scan on Albums <Row> (Full scan, scan_method: Automatic)         |
|  13 |             +- [Map] Local Distributed Union <Row>                                          |
|  14 |                +- Filter Scan <Row> (seekable_key_size: 0)                                  |
| *15 |                   +- Table Scan on Albums <Row> (scan_method: Row)                          |
+-----+---------------------------------------------------------------------------------------------+
Predicates(identified by ID):
 15: Seek Condition: ($SingerId_1 = $sort_SingerId)
```

## v4: secondary index selection for leading indexed column predicates

Official item: version 4 improved secondary index selection, including
preferring secondary indexes with predicates on leading indexed columns even
when optimizer statistics are unavailable or report the base table as small.

The observed example below uses a predicate on `FirstName`, the leading column
of `SingersByFirstLastName`. Version 3 scans the `Singers` base table and keeps
both predicates as residual filters. Version 4 uses
`SingersByFirstLastName` for the `FirstName` split/seek condition, then performs
a distributed apply/back join to the base table for the `BirthDate` residual
filter.

Spanner Omni 2026.r1-beta recheck: with a minimal `Singers` table plus
`SingersByFirstLastName` index, both literal-value and parameterized variants
planned as the v3-style base-table full scan in optimizer versions 3 and 4.
The DBaaS before/after below therefore remains valuable upstream evidence, but
it is not currently reproducible in the empty synthetic Omni schema.

```text
=== optimizer-version/v3/leading-index-predicate ===
@{OPTIMIZER_VERSION=3}
SELECT FirstName FROM Singers
WHERE FirstName = @firstName AND BirthDate BETWEEN @begin AND @end;
+----+-----------------------------------------------------------------------------+
| ID | Operator                                                                    |
+----+-----------------------------------------------------------------------------+
|  0 | Distributed Union on Singers <Row>                                          |
|  1 | +- Local Distributed Union <Row>                                            |
|  2 |    +- Serialize Result <Row>                                                |
| *3 |       +- Filter Scan <Row> (seekable_key_size: 0)                           |
|  4 |          +- Table Scan on Singers <Row> (Full scan, scan_method: Automatic) |
+----+-----------------------------------------------------------------------------+
Predicates(identified by ID):
 3: Residual Condition: (($FirstName = @firstname) AND BETWEEN($BirthDate, @begin, @end))


=== optimizer-version/v4/leading-index-predicate ===
@{OPTIMIZER_VERSION=4}
SELECT FirstName FROM Singers
WHERE FirstName = @firstName AND BirthDate BETWEEN @begin AND @end;
+-----+---------------------------------------------------------------------------------------+
| ID  | Operator                                                                              |
+-----+---------------------------------------------------------------------------------------+
|  *0 | Distributed Union on SingersByFirstLastName <Row>                                     |
|  *1 | +- Distributed Cross Apply <Row>                                                      |
|   2 |    +- [Input] Create Batch <Batch>                                                    |
|   3 |    |  +- RowToDataBlock                                                               |
|   4 |    |     +- Local Distributed Union <Row>                                             |
|  *5 |    |        +- Filter Scan <Row> (seekable_key_size: 0)                               |
|  *6 |    |           +- Index Scan on SingersByFirstLastName <Row> (scan_method: Automatic) |
|  18 |    +- [Map] Serialize Result <Row>                                                    |
|  19 |       +- Cross Apply <Row>                                                            |
|  20 |          +- [Input] KeyRangeAccumulator <Row>                                         |
|  21 |          |  +- DataBlockToRow                                                         |
|  22 |          |     +- Batch Scan on $v2 <Batch> (scan_method: Batch)                      |
|  25 |          +- [Map] Local Distributed Union <Row>                                       |
| *26 |             +- Filter Scan <Row> (seekable_key_size: 0)                               |
| *27 |                +- Table Scan on Singers <Row> (scan_method: Row)                      |
+-----+---------------------------------------------------------------------------------------+
Predicates(identified by ID):
  0: Split Range: ($FirstName = @firstname)
  1: Split Range: ($Singers_key_SingerId'3 = $Singers_key_SingerId'2)
  5: Residual Condition: ($FirstName = @firstname)
  6: Seek Condition: IS_NOT_DISTINCT_FROM($FirstName, @firstname)
 26: Residual Condition: BETWEEN($BirthDate, @begin, @end)
 27: Seek Condition: ($Singers_key_SingerId'3 = $batched_Singers_key_SingerId'3)
```

## v5: cost-based join algorithm selection

Official item: version 5 added cost-based join algorithm selection and join
commutativity. The synthetic example changes from a distributed apply/back-join
shape to a hash join with `BloomFilterBuild`.

```text
=== optimizer-version/v4/distributed/distributed-apply ===
@{OPTIMIZER_VERSION=4}
SELECT AlbumTitle FROM Songs JOIN Albums ON Albums.AlbumId = Songs.AlbumId
+-----+-------------------------------------------------------------------------------------------------------+
| ID  | Operator                                                                                              |
+-----+-------------------------------------------------------------------------------------------------------+
|   0 | Distributed Union on SongsBySingerAlbumSongNameDesc <Row>                                             |
|  *1 | +- Distributed Cross Apply <Row>                                                                      |
|   2 |    +- [Input] Create Batch <Batch>                                                                    |
|   3 |    |  +- RowToDataBlock                                                                               |
|   4 |    |     +- Local Distributed Union <Row>                                                             |
|   5 |    |        +- Index Scan on SongsBySingerAlbumSongNameDesc <Row> (Full scan, scan_method: Automatic) |
|   8 |    +- [Map] Serialize Result <Row>                                                                    |
|   9 |       +- Cross Apply <Row>                                                                            |
|  10 |          +- [Input] DataBlockToRow                                                                    |
|  11 |          |  +- Batch Scan on $v2 <Batch> (scan_method: Batch)                                         |
|  14 |          +- [Map] Local Distributed Union <Row>                                                       |
| *15 |             +- Filter Scan <Row> (seekable_key_size: 0)                                               |
|  16 |                +- Index Scan on AlbumsByAlbumTitle <Row> (Full scan, scan_method: Row)                |
+-----+-------------------------------------------------------------------------------------------------------+
Predicates(identified by ID):
  1: Split Range: ($AlbumId_1 = $AlbumId)
 15: Residual Condition: ($AlbumId_1 = $batched_AlbumId')


=== optimizer-version/v5/distributed/distributed-apply ===
@{OPTIMIZER_VERSION=5}
SELECT AlbumTitle FROM Songs JOIN Albums ON Albums.AlbumId = Songs.AlbumId
+-----+-------------------------------------------------------------------------------------------------------+
| ID  | Operator                                                                                              |
+-----+-------------------------------------------------------------------------------------------------------+
|   0 | Serialize Result <Row>                                                                                |
|  *1 | +- Hash Join <Row> (join_type: INNER)                                                                 |
|   2 |    +- [Build] BloomFilterBuild <Row>                                                                  |
|   3 |    |  +- Distributed Union on SongsBySingerAlbumSongNameDesc <Row>                                    |
|   4 |    |     +- Local Distributed Union <Row>                                                             |
|   5 |    |        +- Index Scan on SongsBySingerAlbumSongNameDesc <Row> (Full scan, scan_method: Automatic) |
|   9 |    +- [Probe] Distributed Union on AlbumsByAlbumTitle <Row>                                           |
|  10 |       +- Local Distributed Union <Row>                                                                |
| *11 |          +- Filter Scan <Row> (seekable_key_size: 0)                                                  |
|  12 |             +- Index Scan on AlbumsByAlbumTitle <Row> (Full scan, scan_method: Automatic)             |
+-----+-------------------------------------------------------------------------------------------------------+
Predicates(identified by ID):
  1: Condition: ($AlbumId_1 = $AlbumId)
 11: Residual Condition: BLOOM_FILTER_MATCH($existence_filter, $AlbumId_1)
```

## v6: full outer join predicate and LIMIT pushdown

Official item: version 6 improved limit pushing and predicate pushing through
full outer joins.

```text
=== optimizer-version/v5/optimizer-gaps/v6/full-outer-join-predicate-limit ===
@{OPTIMIZER_VERSION=5}
SELECT a.SingerId, a.AlbumTitle, s.SongName
FROM Albums AS a
FULL OUTER JOIN Songs AS s
ON a.SingerId = s.SingerId AND a.AlbumId = s.AlbumId
WHERE a.ReleaseDate >= DATE '2020-01-01' OR s.Duration > 180
LIMIT 10
+-----+---------------------------------------------------------------------------------------+
| ID  | Operator                                                                              |
+-----+---------------------------------------------------------------------------------------+
|   0 | Global Limit <Row>                                                                    |
|   1 | +- Distributed Union on Albums <Row> (split_ranges_aligned)                           |
|   2 |    +- Serialize Result <Row>                                                          |
|   3 |       +- Local Limit <Row>                                                            |
|   4 |          +- Local Distributed Union <Row>                                             |
|  *5 |             +- Filter <Row>                                                           |
|   6 |                +- Outer Apply <Row>                                                   |
|   7 |                   +- [Input] Table Scan on Albums <Row> (Full scan, scan_method: Row) |
|  12 |                   +- [Map] Local Distributed Union <Row>                              |
|  13 |                      +- Filter Scan <Row> (seekable_key_size: 0)                      |
| *14 |                         +- Table Scan on Songs <Row> (scan_method: Row)               |
+-----+---------------------------------------------------------------------------------------+
Predicates(identified by ID):
  5: Condition: (($ReleaseDate >= 18262 unix days (2020-01-01)) OR ($Duration_1 > 180))
 14: Seek Condition: (($SingerId_1 = $SingerId) AND ($AlbumId_1 = $AlbumId))


=== optimizer-version/v6/optimizer-gaps/v6/full-outer-join-predicate-limit ===
@{OPTIMIZER_VERSION=6}
SELECT a.SingerId, a.AlbumTitle, s.SongName
FROM Albums AS a
FULL OUTER JOIN Songs AS s
ON a.SingerId = s.SingerId AND a.AlbumId = s.AlbumId
WHERE a.ReleaseDate >= DATE '2020-01-01' OR s.Duration > 180
LIMIT 10
+-----+-----------------------------------------------------------------------------------------------------------+
| ID  | Operator                                                                                                  |
+-----+-----------------------------------------------------------------------------------------------------------+
|   0 | Global Limit <Row>                                                                                        |
|   1 | +- Distributed Union on AlbumsByReleaseDateTitleDesc <Row>                                                |
|   2 |    +- Serialize Result <Row>                                                                              |
|   3 |       +- Local Limit <Row>                                                                                |
|  *4 |          +- Filter <Row>                                                                                  |
|  *5 |             +- Distributed Outer Apply <Row>                                                              |
|   6 |                +- [Input] Create Batch <Row>                                                              |
|   7 |                |  +- Local Distributed Union <Row>                                                        |
|   8 |                |     +- Compute Struct <Row>                                                              |
|   9 |                |        +- Index Scan on AlbumsByReleaseDateTitleDesc <Row> (Full scan, scan_method: Row) |
|  20 |                +- [Map] Cross Apply <Row>                                                                 |
|  21 |                   +- [Input] KeyRangeAccumulator <Row>                                                    |
|  22 |                   |  +- Batch Scan on $v2 <Row> (scan_method: Row)                                        |
|  28 |                   +- [Map] Local Distributed Union <Row>                                                  |
|  29 |                      +- Filter Scan <Row> (seekable_key_size: 0)                                          |
| *30 |                         +- Table Scan on Songs <Row> (scan_method: Row)                                   |
+-----+-----------------------------------------------------------------------------------------------------------+
Predicates(identified by ID):
  4: Condition: (($ReleaseDate' >= 18262 unix days (2020-01-01)) OR ($Duration_1 > 180))
  5: Split Range: (($SingerId_1 = $SingerId) AND ($AlbumId_1 = $AlbumId))
 30: Seek Condition: (($SingerId_1 = $batched_SingerId) AND ($AlbumId_1 = $batched_AlbumId))
```

## v7: cost-based index union

Official item: version 7 added support for cost-based selection of index union
plans. The synthetic OR predicate changes from a single index scan to `Union
All` over two index scans plus `Hash Aggregate`.

```text
=== optimizer-version/v6/optimizer-gaps/v7/unhinted-index-union-candidate ===
@{OPTIMIZER_VERSION=6}
SELECT s.SingerId FROM Singers AS s WHERE s.FirstName = 'Alice' OR s.LastName = 'Smith'
+----+---------------------------------------------------------------------------+
| ID | Operator                                                                  |
+----+---------------------------------------------------------------------------+
| *0 | Distributed Union on SingersByFirstLastName <Row>                         |
|  1 | +- Local Distributed Union <Row>                                          |
|  2 |    +- Serialize Result <Row>                                              |
|  3 |       +- Filter Scan <Row> (seekable_key_size: 2)                         |
| *4 |          +- Index Scan on SingersByFirstLastName <Row> (scan_method: Row) |
+----+---------------------------------------------------------------------------+
Predicates(identified by ID):
 0: Split Range: (($FirstName = 'Alice') OR ($LastName = 'Smith'))
 4: Seek Condition: (($FirstName = 'Alice') OR ($LastName = 'Smith'))


=== optimizer-version/v7/optimizer-gaps/v7/unhinted-index-union-candidate ===
@{OPTIMIZER_VERSION=7}
SELECT s.SingerId FROM Singers AS s WHERE s.FirstName = 'Alice' OR s.LastName = 'Smith'
+-----+------------------------------------------------------------------------------------------+
| ID  | Operator                                                                                 |
+-----+------------------------------------------------------------------------------------------+
|   0 | Serialize Result <Row>                                                                   |
|   1 | +- Hash Aggregate <Row>                                                                  |
|   2 |    +- Union All <Row>                                                                    |
|   3 |       +- Union Input                                                                     |
|  *4 |       |  +- Distributed Union on SingersByFirstLastName <Row>                            |
|   5 |       |     +- Local Distributed Union <Row>                                             |
|   6 |       |        +- Filter Scan <Row> (seekable_key_size: 0)                               |
|  *7 |       |           +- Index Scan on SingersByFirstLastName <Row> (scan_method: Automatic) |
|  18 |       +- Union Input                                                                     |
| *19 |          +- Distributed Union on SingersByLastName <Row>                                 |
|  20 |             +- Local Distributed Union <Row>                                             |
|  21 |                +- Filter Scan <Row> (seekable_key_size: 0)                               |
| *22 |                   +- Index Scan on SingersByLastName <Row> (scan_method: Automatic)      |
+-----+------------------------------------------------------------------------------------------+
Predicates(identified by ID):
  4: Split Range: ($FirstName' = 'Alice')
  7: Seek Condition: IS_NOT_DISTINCT_FROM($FirstName', 'Alice')
 19: Split Range: ($LastName'2 = 'Smith')
 22: Seek Condition: IS_NOT_DISTINCT_FROM($LastName'2, 'Smith')
```

## v8: informational foreign-key handling

Official item: version 8 includes more efficient handling of foreign keys. With
`USE_UNENFORCED_FOREIGN_KEY=TRUE`, v8 removes the referenced-table lookup in
this synthetic schema.

```text
=== optimizer-version/v7/optimizer-gaps/v8/use-unenforced-foreign-key-true ===
@{OPTIMIZER_VERSION=7, USE_UNENFORCED_FOREIGN_KEY=TRUE}
SELECT o.CustomerId
FROM FKOrders AS o
JOIN FKCustomers AS c ON c.CustomerId = o.CustomerId
+-----+---------------------------------------------------------------------------------+
| ID  | Operator                                                                        |
+-----+---------------------------------------------------------------------------------+
|   0 | Distributed Union on FKOrders <Row>                                             |
|  *1 | +- Distributed Cross Apply <Row>                                                |
|   2 |    +- [Input] Create Batch <Batch>                                              |
|   3 |    |  +- RowToDataBlock                                                         |
|   4 |    |     +- Local Distributed Union <Row>                                       |
|   5 |    |        +- Table Scan on FKOrders <Row> (Full scan, scan_method: Automatic) |
|   8 |    +- [Map] Serialize Result <Row>                                              |
|   9 |       +- Cross Apply <Row>                                                      |
|  10 |          +- [Input] KeyRangeAccumulator <Row>                                   |
|  11 |          |  +- DataBlockToRow                                                   |
|  12 |          |     +- Batch Scan on $v2 <Batch> (scan_method: Batch)                |
|  15 |          +- [Map] Local Distributed Union <Row>                                 |
|  16 |             +- Filter Scan <Row> (seekable_key_size: 0)                         |
| *17 |                +- Table Scan on FKCustomers <Row> (scan_method: Row)            |
+-----+---------------------------------------------------------------------------------+
Predicates(identified by ID):
  1: Split Range: ($CustomerId_1 = $CustomerId)
 17: Seek Condition: ($CustomerId_1 = $batched_CustomerId')


=== optimizer-version/v8/optimizer-gaps/v8/use-unenforced-foreign-key-true ===
@{OPTIMIZER_VERSION=8, USE_UNENFORCED_FOREIGN_KEY=TRUE}
SELECT o.CustomerId
FROM FKOrders AS o
JOIN FKCustomers AS c ON c.CustomerId = o.CustomerId
+----+---------------------------------------------------------------------------+
| ID | Operator                                                                  |
+----+---------------------------------------------------------------------------+
|  0 | Distributed Union on FKOrders <Row>                                       |
|  1 | +- Local Distributed Union <Row>                                          |
|  2 |    +- Serialize Result <Row>                                              |
|  3 |       +- Table Scan on FKOrders <Row> (Full scan, scan_method: Automatic) |
+----+---------------------------------------------------------------------------+
```
