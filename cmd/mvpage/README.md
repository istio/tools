# mvpage

This tool simplifies the task of moving markdown pages within istio.io. You run the tool with

```bash
mvpage <original markdown file> <new markdown file>
```

When the original and new markdown files are located in one of Hugo's content directory
within the <https://github.com/istio/istio.io> repo. The command will:

- Move the original file to the new location (creating any necessary directories along the way).

- Update all links to the moved page within the web site to point to the page's new location.

- Add an alias entry to the page's front-matter such that any bookmarks set to the old page's
address will continue working and be redirected to the new address automatically.
