# Kube OpenAPI Integration Tests

## Running the integration tests

Within the current directory:

```bash
$ go test -v .
```

## Generating the golden Swagger definition file

First, run the generator to create `openapi_generated.go` file which specifies
the `OpenAPIDefinition` for each type.

```bash
$ go run ../../cmd/openapi-gen/openapi-gen.go "./testdata/listtype"
```
The generated file `pkg/generaged/openapi_generated.go` should have been created.

Next, run the OpenAPI builder to create the Swagger file which includes
the definitions. The output file named `golden.json` will be output in
the current directory.

```bash
$ go run builder/main.go golden.json
```

