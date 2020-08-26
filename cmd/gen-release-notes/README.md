# Release notes generation

The tooling in this directory is used to generate release notes based on the
[release notes
schema](https://github.com/istio/istio/tree/master/releasenotes). In this
release notes system, a release notes file is created in each pull request which
should have release notes. These notes are then collated in order to generate a
single release notes file.

## Generating Release Notes

To generate release notes, run:

```bash
go build
./gen-release-notes --notes <notes-dir> --templates <templates-dir> --oldBranch myOldBranch --newBranch myNewBranch
```

### Arguments

* (optional) `--notes`  --  indicates where release notes should be found. Default: `./templates`
* (optional) `--templates` -- indicates where templates should be found. Default: `./notes`
* (optional) `--validateOnly` -- indicates to perform validation but not release notes generation.
* `--oldBranch` -- indicates the branch (or tag) to compare against
* `--(newBranch)` -- indicates the branch (or tag) containing new release notes

## Templates

Release notes templates are standard markdown files containing HTML comments
indicating where content should be substituted. These are stored in the
[templates](./templates) directory.

* Release notes can be substituted using:

```html
<!-- releaseNotes -->
```

* Security notes can be substituted using:

```html
<!-- securityNotes -->
```

* Upgrade notes can be substituted using:

```html
<!-- upgradeNotes -->
```

### Filtering notes

Notes can be substituted based on fields in the release notes files.

To substitute for release notes matching the `traffic-management` area, use the
following comment:

```html
<!-- releaseNotes area:traffic-management -->
```
