package main

var cteQueries = []queryCase{
	{
		Label: "cte/constant-single-reference",
		SQL:   `WITH CTE AS (SELECT 1 AS PK, "foo" AS col) SELECT * FROM CTE`,
	},
	{
		Label: "cte/constant-repeated-reference",
		SQL:   `WITH CTE AS (SELECT 1 AS PK, "foo" AS col) SELECT * FROM CTE c1 JOIN CTE c2 USING (PK)`,
	},
	{
		Label: "cte/deterministic-function-single-reference",
		SQL:   `WITH CTE AS (SELECT 1 AS PK, SHA256("foo") AS v) SELECT * FROM CTE`,
	},
	{
		Label: "cte/deterministic-function-repeated-reference",
		SQL:   `WITH CTE AS (SELECT 1 AS PK, SHA256("foo") AS v) SELECT c1.PK, c1.v, c2.v FROM CTE c1 JOIN CTE c2 USING (PK)`,
	},
	{
		Label: "cte/current-timestamp-single-reference",
		SQL:   `WITH CTE AS (SELECT 1 AS PK, CURRENT_TIMESTAMP() AS v) SELECT * FROM CTE`,
	},
	{
		Label: "cte/current-timestamp-repeated-reference",
		SQL:   `WITH CTE AS (SELECT 1 AS PK, CURRENT_TIMESTAMP() AS v) SELECT c1.PK, c1.v, c2.v FROM CTE c1 JOIN CTE c2 USING (PK)`,
	},
	{
		Label: "cte/table-single-reference",
		SQL:   `WITH CTE AS (SELECT SingerId, FirstName FROM Singers) SELECT * FROM CTE WHERE SingerId = 1`,
	},
	{
		Label: "cte/table-repeated-reference",
		SQL:   `WITH CTE AS (SELECT SingerId, FirstName FROM Singers) SELECT c1.SingerId, c1.FirstName, c2.FirstName FROM CTE c1 JOIN CTE c2 USING (SingerId)`,
	},
}
