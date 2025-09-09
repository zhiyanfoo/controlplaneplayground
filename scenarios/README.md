# Test Scenarios

Each of these tests are independent

1. Basic
1. Change referenced cluster in the virtual host
1. Change in cluster definition (Change the loadbalancing policy)
1. Change in listener definition (Change the port the it listens on)
1. Test deleting the referenced cluster (after a successful request is made)
1. Test deleting the referenced virtual host (after a successful request is made)
1. Test where the odcds config is set as typed filter config on the route, and the route is
   changed. This will lead to a crash in envoy due to a bug.
