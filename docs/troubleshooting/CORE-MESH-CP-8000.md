## CORE-MESH-CP-8000
### Message text
Control-plane now works in Phantom mode. Read operation available only.

### Scenario
Database availability issue.

### Reason
The database is currently unavailable.

### Solution
In case of problems with the database, the control-plane switches to read-only mode. Until the database issues are resolved, write operations are not available.