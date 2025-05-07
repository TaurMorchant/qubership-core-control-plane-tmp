## CORE-MESH-CP-8001
### Message text
Database operation failed.

### Scenario
Attempt to write data to the database.

### Reason
The reason should be looked for in the logs of the control-plane.

### Solution
This is unexpected behavior. In most cases, this is an unexpected issue with database availability. At the moment, 
the control plane has not yet switched to read-only mode ([CORE-MESH-CP-8000](./CORE-MESH-CP-8000.md)).