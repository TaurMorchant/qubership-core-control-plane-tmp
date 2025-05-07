# Rate Limiting in Gateways

This article covers rate limiting in Service Mesh gateways. 


[[_TOC_]]

## Overview

You can configure rate limit by setting amount of allowed requests per second for any route, or for the entire virtualService (in terms of control-plane v3 API). All the requests above the allowed limit will be rejected by gateway with HTTP status code 429. 

Please note, that per-virtualService rate limit configuration is forbidden for public, private and internal gateways:  these 3 gateways only support per-route rate limiting. 

## How to set up rate limit
To set up rate limit you need to register two configurations in control-plane (order does not matter): 

1. RateLimit configuration with the unique RateLimit config name;
2. RouteConfiguration with virtualService or route that refers to the RateLimit configuration by its name. 

There are several ways to register both RateLimit and RouteConfiguration: 

1. [Control-Plane REST API](../api/control-plane-api.md)
2. [Declarative API](./development-guide.md#routes-registration-using-configuration-files)

Below are some examples of declarative rate limit configuration.  

## Example

Here is an example of `routes-configuration.yaml` with explanation:

```yaml
---
apiVersion: nc.core.mesh/v3
kind: RateLimit
spec:
  name: ${ENV_SERVICE_NAME}-default-rate-limit
  priority: PROJECT
  limitRequestPerSecond: 100
---
apiVersion: nc.core.mesh/v3
kind: RateLimit
spec:
  name: ${ENV_SERVICE_NAME}-catalog-rate-limit
  priority: PROJECT
  limitRequestPerSecond: 50
---
apiVersion: nc.core.mesh/v3
kind: RateLimit
spec:
  name: ${ENV_SERVICE_NAME}-orders-post-put-rate-limit
  priority: PROJECT
  limitRequestPerSecond: 4
---
apiVersion: nc.core.mesh/v3
kind: RateLimit
spec:
  name: ${ENV_SERVICE_NAME}-orders-get-rate-limit
  priority: PROJECT
  limitRequestPerSecond: 10
---
apiVersion: nc.core.mesh/v3
kind: RouteConfiguration
metadata:
  name: ${ENV_SERVICE_NAME}-routes
  namespace: "${ENV_NAMESPACE}"
spec:
  gateways: ["tbapi-gateway"]
  virtualServices:
  - name: "${ENV_SERVICE_NAME}"
    hosts: ["${ENV_SERVICE_NAME}"]
    rateLimit: "${ENV_SERVICE_NAME}-default-rate-limit"
    routeConfiguration:
      version: "${ENV_DEPLOYMENT_VERSION}"
      routes:
      - destination:
          cluster: "${ENV_SERVICE_NAME}"
          endpoint: http://${DEPLOYMENT_RESOURCE_NAME}:8080
        rules:
        - match:
            prefix: /api/{version}/tbapi/catalogManagement
          prefixRewrite: /api/{version}/catalogManagement
          rateLimit: "${ENV_SERVICE_NAME}-catalog-rate-limit"
        - match:
            prefix: /api/{version}/tbapi/catalogExport
          prefixRewrite: /api/{version}/catalogExport
          rateLimit: "${ENV_SERVICE_NAME}-catalog-rate-limit"
        - match:
            prefix: /api/{version}/tbapi/orderManagement
            headerMatchers:
            - name: ":method"
              safeRegexMatch: "^(POST|Post|post|PUT|Put|put)$"
          prefixRewrite: /api/{version}/orderManagement
          rateLimit: "${ENV_SERVICE_NAME}-orders-post-put-rate-limit"
        - match:
            prefix: /api/{version}/tbapi/orderManagement
            headerMatchers:
            - name: ":method"
              exactMatch: "GET"
          prefixRewrite: /api/{version}/orderManagement
          rateLimit: "${ENV_SERVICE_NAME}-orders-get-rate-limit"
        - match:
            prefix: /tbapi/swagger-ui/
        - match:
            prefix: /tbapi/v3/api-docs
```

In this example several different rate limits configured for different routes:
* Routes `/api/{version}/tbapi/catalogManagement` and `/api/{version}/tbapi/catalogExport` share limit of `50` requests per second, which means that overall number of requests to these two routes each second will be not greater than `50`. 
* Route `/api/{version}/orderManagement` with `PUT` and `POST` method matcher has rate limit of `4` requests per second, which means that overall number of `PUT` and `POST` requests to this endpoint each second will be not greater than `4`. 
* All the other requests to route `/api/{version}/orderManagement` (methods `GET`, `DELETE`, etc.) will be limited by `10` requests per second overall. 
* Requests to `/tbapi/swagger-ui/` and `/tbapi/v3/api-docs` do not have per-route RateLimit configuration, so they will be limited by the RateLimit config on virtualService level, and overall number of requests to these two routes  each second will be not greater than `100`. 
