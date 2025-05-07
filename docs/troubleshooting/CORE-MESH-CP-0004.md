## CORE-MESH-CP-0004
### Message text
Composite operation is forbidden.

### Scenario
There are 2 scenarios in which this error can appear:
1) The client tries to add namespace to composite platform that is already a baseline.
2) The client tries to delete baseline namespace from composite platform.

### Reason
This operation cannot be performed on the baseline namespace.

### Solution
Check if the namespace in the request is correct.
