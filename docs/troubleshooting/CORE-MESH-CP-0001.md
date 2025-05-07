## CORE-MESH-CP-0001
### Message text
BlueGreen operation is forbidden.

### Scenario
The client tries to do one of the following:
1) Promote non-candidate version
2) Deleting non-candidate version
3) Rollback legacy version

### Reason
The client is trying to perform a bluegreen operation that does not match the current state of the cloud. The specific error will be indicated in the message.

### Solution
1) Promote/Deleting non-candidate version - need to make sure that the candidate version was successfully installed on the cloud 
2) Rollback legacy version - it is impossible to rollback the legacy version
