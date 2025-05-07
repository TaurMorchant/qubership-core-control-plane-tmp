FROM alpine:3.19.7

EXPOSE 8080 8443 15010
COPY --chown=10001:0 control-plane/bin/control-plane-service /app/control-plane
COPY --chown=10001:0 ["control-plane/application.yaml", "control-plane/api-version-info.json", "control-plane/constancy/migration/*.sql", "control-plane/docs", "/app/"]