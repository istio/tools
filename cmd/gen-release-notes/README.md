# Release notes generation

The tooling in this directory is used to generate release notes based on the
[release notes
schema](https://github.com/istio/istio/tree/master/releasenotes). In this
release notes system, a release notes file is created in each pull request which
should have release notes. These notes are then collated in order to generate a
single release notes file.

## Generating Release Notes

If both Istio and tools are cloned in the same directory and you want to generate release notes for changes in the `release-1.7` branch since the `1.7.0` tag was created, you could run the following from the `tools/cmd/gen-release-notes` directory:

```bash
pushd ../../../istio/releasenotes/notes
git checkout release-1.7
popd

go build
./gen-release-notes --notes ../../../istio/releasenotes/notes --oldBranch 1.7.0 --newBranch release-1.7
```

### Arguments

* (optional) `--notes`  --  indicates where release notes should be found. Default: `./notes`. This argument can be repeated for additional repositories.
* (optional) `--templates` -- indicates where templates should be found. Default: `./templates`
* (optional) `--validateOnly` -- indicates to perform validation but not release notes generation.
* `--oldBranch` -- indicates the branch (or tag) to compare against
* `--newBranch` -- indicates the branch (or tag) containing new release notes
* `--oldRelease` -- indicates the name of the release being upgrade from
* `--newRelease` -- indicates the name of the new release.

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
