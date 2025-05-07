## CORE-MESH-CP-4001
### Message text
Validation error.

### Scenario
The client tries to any request to the control-plane.

### Reason
In this case, the reason is given in details of error.

### Solution
To resolve this error, the client should look at the details of the error. Error details should contain reason of what the request did not pass validation.

For example
```json
{
    "id": "fdaaad8a-e010-11ed-b4be-00ac4f56bd8b",
    "code": "CORE-MESH-CP-4001",
    "reason": "Validation error",
    "message": "Namespace path variable must no be empty",
    "status": "400",
    "@type": "NC.TMFErrorResponse.v1.0"
}
```