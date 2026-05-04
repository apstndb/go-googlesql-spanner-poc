package spanalyzer

import "cloud.google.com/go/spanner/apiv1/spannerpb"

const spannerSysName = "SPANNER_SYS"

type spannerSysColumn struct {
	name string
	typ  *TypeSpec
}

type spannerSysTable struct {
	name    string
	columns []spannerSysColumn
}

func (c *Catalog) addSpannerSysTables() {
	for _, def := range spannerSysTables {
		table := &Table{
			Name:    ObjectName{Parts: []string{spannerSysName, def.name}},
			Columns: make([]*Column, 0, len(def.columns)),
		}
		for _, column := range def.columns {
			table.Columns = append(table.Columns, &Column{
				Name: column.name,
				Type: column.typ,
			})
		}
		c.Tables[table.Name.String()] = table
	}
}

var (
	sysDistributionType = sysArray(sysStruct(
		sysField("COUNT", sysInt64Type()),
		sysField("MEAN", sysFloat64Type()),
		sysField("SUM_OF_SQUARED_DEVIATION", sysFloat64Type()),
		sysField("NUM_FINITE_BUCKETS", sysInt64Type()),
		sysField("GROWTH_FACTOR", sysFloat64Type()),
		sysField("SCALE", sysFloat64Type()),
		sysField("BUCKET_COUNTS", sysArray(sysInt64Type())),
	))
	sysLockRequestsType = sysArray(sysStruct(
		sysField("column", sysStringType()),
		sysField("lock_mode", sysStringType()),
		sysField("transaction_tag", sysStringType()),
	))
	sysOperationsByTableType = sysArray(sysStruct(
		sysField("TABLE", sysStringType()),
		sysField("INSERT_OR_UPDATE_COUNT", sysInt64Type()),
		sysField("INSERT_OR_UPDATE_BYTES", sysInt64Type()),
	))
	sysVectorPercentilesType = sysArray(sysStruct(
		sysField("percentile", sysInt64Type()),
		sysField("value_at_percentile", sysFloat64Type()),
	))
)

var (
	spannerSysIntervalColumn = sysTimestamp("INTERVAL_END")
	spannerSysTopIntervals   = []string{"MINUTE", "10MINUTE", "HOUR"}
)

var spannerSysTables = buildSpannerSysTables()

func buildSpannerSysTables() []spannerSysTable {
	tables := []spannerSysTable{
		sysTable("ACTIVE_PARTITIONED_DMLS",
			sysString("TEXT"),
			sysInt64("TEXT_FINGERPRINT"),
			sysString("SESSION_ID"),
			sysInt64("NUM_PARTITIONS_TOTAL"),
			sysInt64("NUM_PARTITIONS_COMPLETE"),
			sysInt64("NUM_TRIVIAL_PARTITIONS_COMPLETE"),
			sysFloat64("PROGRESS"),
			sysInt64("ROWS_PROCESSED"),
			sysTimestamp("START_TIMESTAMP"),
			sysTimestamp("LAST_UPDATE_TIMESTAMP"),
		),
		sysTable("OLDEST_ACTIVE_QUERIES",
			sysTimestamp("START_TIME"),
			sysInt64("TEXT_FINGERPRINT"),
			sysString("TEXT"),
			sysBool("TEXT_TRUNCATED"),
			sysString("SESSION_ID"),
			sysString("QUERY_ID"),
			sysString("CLIENT_IP_ADDRESS"),
			sysString("API_CLIENT_HEADER"),
			sysString("USER_AGENT_HEADER"),
			sysString("SERVER_REGION"),
			sysString("PRIORITY"),
			sysString("TRANSACTION_TYPE"),
		),
		sysTable("ACTIVE_QUERIES_SUMMARY",
			sysInt64("ACTIVE_COUNT"),
			sysTimestamp("OLDEST_START_TIME"),
			sysInt64("COUNT_OLDER_THAN_1S"),
			sysInt64("COUNT_OLDER_THAN_10S"),
			sysInt64("COUNT_OLDER_THAN_100S"),
		),
		sysTable("TABLE_SIZES_STATS_1HOUR",
			spannerSysIntervalColumn,
			sysString("TABLE_NAME"),
			sysFloat64("USED_BYTES"),
			sysFloat64("USED_SSD_BYTES"),
			sysFloat64("USED_HDD_BYTES"),
		),
		sysTable("TABLE_SIZES_STATS_PER_LOCALITY_GROUP_1HOUR",
			spannerSysIntervalColumn,
			sysString("TABLE_NAME"),
			sysString("LOCALITY_GROUP"),
			sysFloat64("USED_BYTES"),
			sysFloat64("USED_SSD_BYTES"),
			sysFloat64("USED_HDD_BYTES"),
		),
		sysTable("VECTOR_INDEX_STATS",
			sysTimestamp("START_TIME"),
			sysString("VECTOR_INDEX_NAME"),
			sysInt64("NUM_LEAVES"),
			sysInt64("NUM_CLUSTERS_SAMPLED"),
			sysInt64("NUM_ZERO_SIZE_CLUSTERS_SAMPLED"),
			sysColumn("CLUSTER_SIZE_PERCENTILES", sysVectorPercentilesType),
			sysColumn("CLUSTER_AVERAGE_DISTANCE_TO_CENTROID_PERCENTILES", sysVectorPercentilesType),
		),
		sysTable("VECTOR_INDEX_METRICS_HISTORY",
			sysTimestamp("START_TIME"),
			sysString("VECTOR_INDEX_NAME"),
			sysInt64("NUM_LEAVES"),
			sysInt64("NUM_CLUSTERS_SAMPLED"),
			sysInt64("NUM_ZERO_SIZE_CLUSTERS_SAMPLED"),
			sysColumn("CLUSTER_SIZE_PERCENTILES", sysVectorPercentilesType),
			sysColumn("CLUSTER_AVERAGE_DISTANCE_TO_CENTROID_PERCENTILES", sysVectorPercentilesType),
		),
	}

	tables = append(tables, spannerSysIntervalTables("COLUMN_OPERATIONS_STATS", spannerSysTopIntervals, spannerSysColumnOperationsColumns)...)
	tables = append(tables, spannerSysIntervalTables("TABLE_OPERATIONS_STATS", spannerSysTopIntervals, spannerSysTableOperationsColumns)...)
	tables = append(tables, spannerSysIntervalTables("LOCK_STATS_TOP", spannerSysTopIntervals, spannerSysLockTopColumns)...)
	tables = append(tables, spannerSysIntervalTables("LOCK_STATS_TOTAL", spannerSysTopIntervals, spannerSysLockTotalColumns)...)
	tables = append(tables, spannerSysIntervalTables("SPLIT_STATS_TOP", spannerSysTopIntervals, spannerSysSplitTopColumns)...)
	tables = append(tables, spannerSysIntervalTables("READ_STATS_TOP", spannerSysTopIntervals, spannerSysReadTopColumns)...)
	tables = append(tables, spannerSysIntervalTables("READ_STATS_TOTAL", spannerSysTopIntervals, spannerSysReadTotalColumns)...)
	tables = append(tables, spannerSysIntervalTables("QUERY_STATS_TOP", spannerSysTopIntervals, spannerSysQueryTopColumns)...)
	tables = append(tables, spannerSysIntervalTables("QUERY_STATS_TOTAL", spannerSysTopIntervals, spannerSysQueryTotalColumns)...)
	tables = append(tables, spannerSysIntervalTables("TXN_STATS_TOP", spannerSysTopIntervals, spannerSysTxnTopColumns)...)
	tables = append(tables, spannerSysIntervalTables("TXN_STATS_TOTAL", spannerSysTopIntervals, spannerSysTxnTotalColumns)...)

	return tables
}

var spannerSysColumnOperationsColumns = []spannerSysColumn{
	spannerSysIntervalColumn,
	sysString("TABLE_NAME"),
	sysString("COLUMN_NAME"),
	sysInt64("READ_COUNT"),
	sysInt64("QUERY_COUNT"),
	sysInt64("WRITE_COUNT"),
	sysBool("IS_QUERY_CACHE_MEMORY_CAPPED"),
}

var spannerSysTableOperationsColumns = []spannerSysColumn{
	spannerSysIntervalColumn,
	sysString("TABLE_NAME"),
	sysInt64("READ_QUERY_COUNT"),
	sysInt64("WRITE_COUNT"),
	sysInt64("DELETE_COUNT"),
}

var spannerSysLockTopColumns = []spannerSysColumn{
	spannerSysIntervalColumn,
	sysBytes("ROW_RANGE_START_KEY"),
	sysFloat64("LOCK_WAIT_SECONDS"),
	sysColumn("SAMPLE_LOCK_REQUESTS", sysLockRequestsType),
	sysString("SAMPLE_LOCK_REQUESTS_JSON_STRING"),
}

var spannerSysLockTotalColumns = []spannerSysColumn{
	spannerSysIntervalColumn,
	sysFloat64("TOTAL_LOCK_WAIT_SECONDS"),
}

var spannerSysSplitTopColumns = []spannerSysColumn{
	spannerSysIntervalColumn,
	sysString("SPLIT_START"),
	sysString("SPLIT_LIMIT"),
	sysInt64("CPU_USAGE_SCORE"),
	sysColumn("AFFECTED_TABLES", sysArray(sysStringType())),
	sysColumn("UNSPLITTABLE_REASONS", sysArray(sysStringType())),
}

var spannerSysReadTopColumns = []spannerSysColumn{
	spannerSysIntervalColumn,
	sysString("REQUEST_TAG"),
	sysString("READ_TYPE"),
	sysColumn("READ_COLUMNS", sysArray(sysStringType())),
	sysInt64("FPRINT"),
	sysInt64("EXECUTION_COUNT"),
	sysFloat64("AVG_ROWS"),
	sysFloat64("AVG_BYTES"),
	sysFloat64("AVG_CPU_SECONDS"),
	sysFloat64("AVG_LOCKING_DELAY_SECONDS"),
	sysFloat64("AVG_CLIENT_WAIT_SECONDS"),
	sysFloat64("AVG_LEADER_REFRESH_DELAY_SECONDS"),
	sysInt64("RUN_IN_RW_TRANSACTION_EXECUTION_COUNT"),
	sysFloat64("AVG_DISK_IO_COST"),
}

var spannerSysReadTotalColumns = []spannerSysColumn{
	spannerSysIntervalColumn,
	sysInt64("EXECUTION_COUNT"),
	sysFloat64("AVG_ROWS"),
	sysFloat64("AVG_BYTES"),
	sysFloat64("AVG_CPU_SECONDS"),
	sysFloat64("AVG_LOCKING_DELAY_SECONDS"),
	sysFloat64("AVG_CLIENT_WAIT_SECONDS"),
	sysFloat64("AVG_LEADER_REFRESH_DELAY_SECONDS"),
	sysInt64("RUN_IN_RW_TRANSACTION_EXECUTION_COUNT"),
}

var spannerSysQueryTopColumns = []spannerSysColumn{
	spannerSysIntervalColumn,
	sysString("REQUEST_TAG"),
	sysString("QUERY_TYPE"),
	sysString("TEXT"),
	sysBool("TEXT_TRUNCATED"),
	sysInt64("TEXT_FINGERPRINT"),
	sysInt64("EXECUTION_COUNT"),
	sysFloat64("AVG_LATENCY_SECONDS"),
	sysFloat64("AVG_ROWS"),
	sysFloat64("AVG_BYTES"),
	sysFloat64("AVG_ROWS_SCANNED"),
	sysFloat64("AVG_CPU_SECONDS"),
	sysInt64("ALL_FAILED_EXECUTION_COUNT"),
	sysFloat64("ALL_FAILED_AVG_LATENCY_SECONDS"),
	sysInt64("CANCELLED_OR_DISCONNECTED_EXECUTION_COUNT"),
	sysInt64("TIMED_OUT_EXECUTION_COUNT"),
	sysFloat64("AVG_BYTES_WRITTEN"),
	sysFloat64("AVG_ROWS_WRITTEN"),
	sysInt64("STATEMENT_COUNT"),
	sysInt64("RUN_IN_RW_TRANSACTION_EXECUTION_COUNT"),
	sysColumn("LATENCY_DISTRIBUTION", sysDistributionType),
	sysFloat64("AVG_MEMORY_PEAK_USAGE_BYTES"),
	sysFloat64("AVG_MEMORY_USAGE_PERCENTAGE"),
	sysFloat64("AVG_QUERY_PLAN_CREATION_TIME_SECS"),
	sysFloat64("AVG_FILESYSTEM_DELAY_SECS"),
	sysFloat64("AVG_REMOTE_SERVER_CALLS"),
	sysFloat64("AVG_ROWS_SPOOLED"),
	sysFloat64("AVG_DISK_IO_COST"),
	sysString("LATENCY_DISTRIBUTION_JSON_STRING"),
}

var spannerSysQueryTotalColumns = []spannerSysColumn{
	spannerSysIntervalColumn,
	sysInt64("EXECUTION_COUNT"),
	sysFloat64("AVG_LATENCY_SECONDS"),
	sysFloat64("AVG_ROWS"),
	sysFloat64("AVG_BYTES"),
	sysFloat64("AVG_ROWS_SCANNED"),
	sysFloat64("AVG_CPU_SECONDS"),
	sysInt64("ALL_FAILED_EXECUTION_COUNT"),
	sysFloat64("ALL_FAILED_AVG_LATENCY_SECONDS"),
	sysInt64("CANCELLED_OR_DISCONNECTED_EXECUTION_COUNT"),
	sysInt64("TIMED_OUT_EXECUTION_COUNT"),
	sysFloat64("AVG_BYTES_WRITTEN"),
	sysFloat64("AVG_ROWS_WRITTEN"),
	sysInt64("RUN_IN_RW_TRANSACTION_EXECUTION_COUNT"),
	sysColumn("LATENCY_DISTRIBUTION", sysDistributionType),
	sysString("LATENCY_DISTRIBUTION_JSON_STRING"),
}

var spannerSysTxnTopColumns = []spannerSysColumn{
	spannerSysIntervalColumn,
	sysString("TRANSACTION_TAG"),
	sysInt64("FPRINT"),
	sysColumn("READ_COLUMNS", sysArray(sysStringType())),
	sysColumn("WRITE_CONSTRUCTIVE_COLUMNS", sysArray(sysStringType())),
	sysColumn("WRITE_DELETE_TABLES", sysArray(sysStringType())),
	sysInt64("ATTEMPT_COUNT"),
	sysInt64("COMMIT_ATTEMPT_COUNT"),
	sysInt64("COMMIT_ABORT_COUNT"),
	sysInt64("COMMIT_RETRY_COUNT"),
	sysInt64("COMMIT_FAILED_PRECONDITION_COUNT"),
	sysInt64("SERIALIZABLE_PESSIMISTIC_TXN_COUNT"),
	sysInt64("SERIALIZABLE_OPTIMISTIC_TXN_COUNT"),
	sysInt64("REPEATABLE_READ_OPTIMISTIC_TXN_COUNT"),
	sysFloat64("AVG_PARTICIPANTS"),
	sysFloat64("AVG_TOTAL_LATENCY_SECONDS"),
	sysFloat64("AVG_COMMIT_LATENCY_SECONDS"),
	sysFloat64("AVG_BYTES"),
	sysColumn("TOTAL_LATENCY_DISTRIBUTION", sysDistributionType),
	sysColumn("OPERATIONS_BY_TABLE", sysOperationsByTableType),
	sysString("TOTAL_LATENCY_DISTRIBUTION_JSON_STRING"),
	sysString("OPERATIONS_BY_TABLE_JSON_STRING"),
}

var spannerSysTxnTotalColumns = []spannerSysColumn{
	spannerSysIntervalColumn,
	sysInt64("ATTEMPT_COUNT"),
	sysInt64("COMMIT_ATTEMPT_COUNT"),
	sysInt64("COMMIT_ABORT_COUNT"),
	sysInt64("COMMIT_RETRY_COUNT"),
	sysInt64("COMMIT_FAILED_PRECONDITION_COUNT"),
	sysInt64("SERIALIZABLE_PESSIMISTIC_TXN_COUNT"),
	sysInt64("REPEATABLE_READ_OPTIMISTIC_TXN_COUNT"),
	sysFloat64("AVG_PARTICIPANTS"),
	sysFloat64("AVG_TOTAL_LATENCY_SECONDS"),
	sysFloat64("AVG_COMMIT_LATENCY_SECONDS"),
	sysFloat64("AVG_BYTES"),
	sysColumn("TOTAL_LATENCY_DISTRIBUTION", sysDistributionType),
	sysColumn("OPERATIONS_BY_TABLE", sysOperationsByTableType),
	sysString("TOTAL_LATENCY_DISTRIBUTION_JSON_STRING"),
	sysString("OPERATIONS_BY_TABLE_JSON_STRING"),
}

func spannerSysIntervalTables(prefix string, intervals []string, columns []spannerSysColumn) []spannerSysTable {
	tables := make([]spannerSysTable, 0, len(intervals))
	for _, interval := range intervals {
		tables = append(tables, sysTable(prefix+"_"+interval, columns...))
	}
	return tables
}

func sysTable(name string, columns ...spannerSysColumn) spannerSysTable {
	return spannerSysTable{name: name, columns: columns}
}

func sysColumn(name string, typ *TypeSpec) spannerSysColumn {
	return spannerSysColumn{name: name, typ: typ}
}

func sysString(name string) spannerSysColumn {
	return sysColumn(name, sysStringType())
}

func sysInt64(name string) spannerSysColumn {
	return sysColumn(name, sysInt64Type())
}

func sysFloat64(name string) spannerSysColumn {
	return sysColumn(name, sysFloat64Type())
}

func sysBool(name string) spannerSysColumn {
	return sysColumn(name, &TypeSpec{Code: spannerpb.TypeCode_BOOL})
}

func sysBytes(name string) spannerSysColumn {
	return sysColumn(name, &TypeSpec{Code: spannerpb.TypeCode_BYTES, Max: true})
}

func sysTimestamp(name string) spannerSysColumn {
	return sysColumn(name, &TypeSpec{Code: spannerpb.TypeCode_TIMESTAMP})
}

func sysStringType() *TypeSpec {
	return &TypeSpec{Code: spannerpb.TypeCode_STRING, Max: true}
}

func sysInt64Type() *TypeSpec {
	return &TypeSpec{Code: spannerpb.TypeCode_INT64}
}

func sysFloat64Type() *TypeSpec {
	return &TypeSpec{Code: spannerpb.TypeCode_FLOAT64}
}

func sysArray(element *TypeSpec) *TypeSpec {
	return &TypeSpec{Code: spannerpb.TypeCode_ARRAY, ArrayElement: element}
}

func sysStruct(fields ...StructField) *TypeSpec {
	return &TypeSpec{Code: spannerpb.TypeCode_STRUCT, StructFields: fields}
}

func sysField(name string, typ *TypeSpec) StructField {
	return StructField{Name: name, Type: typ}
}
