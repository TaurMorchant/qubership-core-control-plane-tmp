package errorcodes

import (
	errs "github.com/netcracker/qubership-core-lib-go-error-handling/v3/errors"
	"net/http"
)

const StatusMasterNodeUnavailable = 521

var OkErrorCode = errs.ErrorCode{Code: "CORE-MESH-CP-0000", Title: "Ok"}

// Business case errors (0001-1499)
var BlueGreenConflictError = cpRestErrorCode{errs.ErrorCode{Code: "CORE-MESH-CP-0001", Title: "BlueGreen operation is forbidden"}, http.StatusConflict, true}
var OperationOnArchivedVersionError = cpRestErrorCode{errs.ErrorCode{Code: "CORE-MESH-CP-0002", Title: "Operation cannot be performed on ARCHIVED version"}, http.StatusConflict, true}
var NotFoundEntityError = cpRestErrorCode{errs.ErrorCode{Code: "CORE-MESH-CP-0003", Title: "Entity not found"}, http.StatusNotFound, true}
var CompositeConflictError = cpRestErrorCode{errs.ErrorCode{Code: "CORE-MESH-CP-0004", Title: "Composite operation is forbidden"}, http.StatusConflict, true}

// Reserved for features errors (2000-2999)
var UnknownErrorCode = errs.ErrorCode{Code: "CORE-MESH-CP-2000", Title: "Unexpected exception"}
var RemoteRequestError = errs.ErrorCode{Code: "CORE-MESH-CP-2001", Title: "Remote request failed"}
var MultiCauseApplyConfigError = cpRestErrorCode{errs.ErrorCode{Code: "CORE-MESH-CP-2002", Title: "Failed to apply one or more configs"}, http.StatusInternalServerError, false}
var ApplyConfigError = errs.ErrorCode{Code: "CORE-MESH-CP-2003", Title: "Can not apply config"}
var MasterNodeError = cpRestErrorCode{errs.ErrorCode{Code: "CORE-MESH-CP-2004", Title: "Master node is not ready yet"}, StatusMasterNodeUnavailable, true}

// Validation rest errors (4000-6999)
var UnmarshalRequestError = cpRestErrorCode{errs.ErrorCode{Code: "CORE-MESH-CP-4000", Title: "Invalid format of the request body"}, http.StatusBadRequest, false}
var ValidationRequestError = cpRestErrorCode{errs.ErrorCode{Code: "CORE-MESH-CP-4001", Title: "Validation error"}, http.StatusBadRequest, false}

// DB errors (8000-8999)
var PhantomModeError = cpRestErrorCode{errs.ErrorCode{Code: "CORE-MESH-CP-8000", Title: "Control-plane now works in Phantom mode. Read operation available only"}, http.StatusServiceUnavailable, false}
var DbOperationError = cpRestErrorCode{errs.ErrorCode{Code: "CORE-MESH-CP-8001", Title: "Database operation failed"}, http.StatusInternalServerError, true}
