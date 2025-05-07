## CORE-MESH-CP-0002
### Message text
Operation cannot be performed on ARCHIVED version.

### Scenario
The client tries to modify stateful session for ARCHIVED deployment version.

### Reason
Cannot modify stateful session configuration for ARCHIVED deployment version.

### Solution
Check if the version in the request is correct. Do not change stateful session configuration for the ARCHIVED version of the deployment.
