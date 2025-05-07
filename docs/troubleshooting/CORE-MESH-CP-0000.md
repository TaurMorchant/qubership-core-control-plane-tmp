## CORE-MESH-CP-0000
### Message text
Ok.

### Scenario
The client tries to apply control-plane configs.

### Reason
Apply configs failed to apply one or more configs. This code will be specified in the configs, which ended in success.

### Solution
This code is only needed to separate apply successful configs from unsuccessful ones.
