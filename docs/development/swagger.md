# How to generate swagger doc manually

1. install swag/cmd v1.8.12 or above
   ```
   go install github.com/swaggo/swag/cmd/swag@latest
   ```
2. from withing `control-plane` folder execute `swag init -g server.go --parseDependency  --parseInternal --parseDepth 1` to generate `swagger.json` file
   ```
   swag init -g server.go --parseDependency  --parseInternal --parseDepth 1
   ```
3. install https://github.com/go-swagger/go-swagger
4. generate MD doc from `swagger.json` file
   ```
   swagger generate markdown -f ./docs/swagger.json --output ./../docs/api/control-plane-api.md
   ```
