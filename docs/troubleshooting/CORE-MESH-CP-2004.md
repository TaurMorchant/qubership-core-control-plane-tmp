## CORE-MESH-CP-2004
### Message text
Master node is not ready yet.

### Scenario
The client tries to send a request to the control-plane when:
1) Control-plane is not deployed yet
2) Master node changes

### Reason
Control-plane currently cannot accept write requests.

### Solution
At the moment, the master node is switching. Need to retry request in a few seconds. If the problem persists after a few seconds, please contact support.