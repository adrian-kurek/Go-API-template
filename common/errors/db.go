package errors

const FailedToPrepareQuery = "failed to prepared statement for execution"

const (
	FailedToExecuteInsertQuery = "failed to execute an insert query"
	FailedToExecuteSelectQuery = "failed to execute select query"
	FailedToExecuteDeleteQuery = "failed to execute a delete query"
	FailedToExecuteUpdateQuery = "failed to execute a update query"
)

const (
	FailedToScanRows        = "failed to scan rows"
	FailedToIterateOverRows = "failed to iterate over rows"
	FailedToScanRow         = "failed to scan rows"
)

const FailedToGetDataFromDatabase = "failed to get data from database"

const (
	FailedToCloseStatement = "failed to close statement"
	FailedToCloseRows      = "failed to close rows"
)
