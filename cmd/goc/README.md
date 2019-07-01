# goc

`goc` is a `go` command-line tool wrapper that injects code-coverage instrumentation into generated code.
It instruments the compiled code using the standard Go toolchain command, and adds boilerplate code that allows
collection of the coverage data through Ctrlz.

`goc` mimics `go` and is meant to be a transparent utility. For all commands, except `build`, it will silently
call `go` internally, as-is, passing all the command-line parameters and environment variables. For `build` command,
it will copy and create and instrumented version of the code, before using the supplied command-line parametes and
environment variables to invoke `go build` on the instrumented code.