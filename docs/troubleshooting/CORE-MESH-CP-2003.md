## CORE-MESH-CP-2003
### Message text
Can not apply config.

### Scenario
The client tries to apply control-plane config. May also be part of CORE-MESH-CP-2002 error code.

### Reason
An error occurred while applying the configuration that will be indicated in the error message.

### Solution
The error details should indicate the reason why apply config failed. 
1) There may be a problem in the request that could not be tracked during the validation stage. It is necessary to check if the configuration is correct - [apply-configuration](../api/control-plane-api.md#apply-configuration). 
2) It is also possible an internal error for which there is no error code

In both cases, the error will contain details that may help solve the problem.