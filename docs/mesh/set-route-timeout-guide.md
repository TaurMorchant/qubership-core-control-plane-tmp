# How to set route timeout

[[_TOC_]]

**WARNING:** This solution only for changing timeouts in runtime. It will be effective until the route is registered again with the other timeout configuration (e.g. old timeout or default timeout (120 seconds for timeout, unset for idle timeout)).
Routes will be registered during pod restart if the service uses libraries to register routes, or during redeploy when registering routes declaratively.

## Best Practice
Default value for timeout (120s) can be considered as recommended. In case you need to significantly increase timeout for some route, please, consider rewriting this long-running operation API in asynchronous style.

## Step 1. Get configuration of route you wish to change.

## Step 2. Apply new route configuration with updated timeout
1. Prepare route configuration body such as next yaml example:
  ```yaml
   apiVersion: nc.core.mesh/v3
   kind: RouteConfiguration
   metadata:
     name: internal-routes
     namespace: ""
   spec:
     gateways: ["internal-gateway-service"]
     virtualServices:
       - name: internal-gateway-service
         routeConfiguration:
           version: "v1"
           routes:
             - destination:
                 cluster: "control-plane"
                 endpoint: control-plane:8080
               rules:
                 - match:
                     prefix: /api/v1/routes
                   prefixRewrite: /api/v1/routes/internal-gateway-service
                   timeout: 600000
                   idleTimeout: 600000
  ```
  
  The fields of this yaml are fill from route received configuration on the [previous step](#step-1-get-configuration-of-route-you-wish-to-change).
  - Explanation fields of the example yaml (spec section):
    - gateways - gateways in which this route configuration should be applied
    - virtualServices.name - Name of the virtual service containing the route. For public, private and internal gateways it equals to gateway name. For facade or composite gateway it usually equals to the microservice family name.
    - virtualServices.routeConfiguration.version - route version (it may be received from route configuration via cpcli (column "**VER**") or Mesh UI)
    - virtualServices.routeConfiguration.routes.destination.cluster - microservice family name (does not contain version suffix)
    - virtualServices.routeConfiguration.routes.destination.endpoint - deployment name and port (it may be received from route configuration via cpcli (column "**ENDPOINT**") or Mesh UI)
    - virtualServices.routeConfiguration.routes.rules.match.prefix - API prefix (it may be received from route configuration via cpcli (column "**PREFIX**") or Mesh UI)
    - virtualServices.routeConfiguration.routes.rules.prefixRewrite - route prefix rewrite (it may be received from route configuration via cpcli (column "**REWRITE**") or Mesh UI)
    - **virtualServices.routeConfiguration.routes.rules.timeout** - optional field, allows to set timeout for route (unit of measurement - milliseconds). Default value is 120 seconds (it is taken from property envoy-proxy.routes.timeout).
    - **virtualServices.routeConfiguration.routes.rules.idleTimeout** - optional field, allows to set idle timeout for route (unit of measurement - milliseconds). By default idle timeout for route is disabled. 

2. If you didn't log in the cpcli system earlier, do the first step (Control Plane CLI login) described [here](#bull-via-control-plane-cli-cpcli).
3. Apply new route configuration with timeout from yaml file with route configuration:
   
   You can apply this configuration via [control-plane REST API](../api/control-plane-api.md#apply-configuration)
   
   Or via CPCLI:
   ```shell script
   cpcli apply -f ./routes-configuration.yaml
   ```

   Result:
   ```
   Configuration RouteConfiguration has applied
   SUCCESS: All configurations successfully applied.
   ```
