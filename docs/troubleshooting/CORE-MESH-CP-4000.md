## CORE-MESH-CP-4000
### Message text
Invalid format of the request body.

### Scenario
The client tries to send request with a body content to the control-plane.

### Reason
Unable to unmarshal request body.

### Solution
This error means that the client sent an invalid request body content and control plane cannot read this body content.

Client should refer to the documentation [control-plane-api](../api/control-plane-api.md) and check if the request was sent correctly.