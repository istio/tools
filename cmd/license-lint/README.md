# Istio License Linter

WARNING: This tool requires <https://github.com/benbalter/licensee> to be in the path.

This tool is used to ensure that the dependencies of a repo have acceptable licenses. The tool can be used in
three modes:

1. Linting. This ensures all dependencies of the current repo have unrestricted or reciprocal licenses, producing an error if
any modules have a restricted or unknown license:

    ```bash
    $ license-lint --config <config file>
    ```

1. Report. Lists license information for all dependencies:

    ```bash
    $ license-lint --config <config file> --report
    ```

1. CSV. Lists license information for all dependencies in CSV format:

    ```bash
    $ license-lint --config <config file> --report
    ```

1. Dumps. Shows all licenses for all dependencies:

    ```bash
    $ license-lint --config <config file> --dump
    ```

The configuration is specified in a YAML file with four stanzas:

```yaml
unrestricted_licenses:
  - license name 1
  - license name 2
reciprocal_licenses:
  - license name 1
  - license name 2
restricted_licenses:
  - license name 1
  - license name 2
whitelisted_modules:
  - module name 1
  - module name 2
```
