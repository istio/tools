
## What's this for?

`protoc-gen-openapi` is a plugin for the Google protocol buffer compiler to generate
openAPI V3 spec for any given input protobuf. It runs as a `protoc-gen-` binary that the
protobuf compiler infers from the `openapi_out` flag.

## Build `protoc-gen-openapi`

`protoc-gen-openapi` is written in Go, so ensure that is installed on your system. You
can follow the instructions on the [golang website](https://golang.org/doc/install) or
on Debian or Ubuntu, you can install it from the package manager:

```bash
sudo apt-get install -y golang
```

To build, first ensure you have the protocol compiler (protoc):

```bash
go get github.com/golang/protobuf/proto
```
To build, run the following command from this project directory:

```bash
go build
```

Then ensure the resulting `protoc-gen-openapi` binary is in your `PATH`. A recommended location
is `$HOME/bin`:

```bash
cp protoc-gen-openapi $HOME/bin
```

Since the following is often in your `$HOME/.bashrc` file:

```bash
export PATH=$HOME/bin:$PATH
```

## Using protoc-gen-openapi

---
**TIP**

The -I option in protoc is useful when you need to specify proto paths for imports.

---

Then to generate the OpenAPI spec of the protobuf defined by file.proto, run

```bash
protoc --openapi_out=output_directory input_directory/file.proto
```

With that input, the output will be written to

	output_directory/file.json

Other supported options are:
*   `per_file`
    *   when set to `true`, the output is per proto file instead of per package.
*   `single_file`
    *   when set to `true`, the output is a single file of all the input protos specified.
*   `use_ref`
    *   when set to `true`, the output uses the `$ref` field in OpenAPI spec to reference other schemas.
*   `yaml`
    *   when set to `true`, the output is in yaml file.