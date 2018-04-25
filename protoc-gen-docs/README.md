
# What's this for?

`protoc-gen-docs` is a plugin for the Google protocol buffer compiler to generate
documentation for any given input protobuf. It runs as a `protoc-gen-` binary that the
protobuf compiler infers from the `docs_out` flag.

## Build `protoc-gen-docs`

`protoc-gen-docs` is written in Go, so ensure that is installed on your system. You
can follow the instructions on the [golang website](https://golang.org/doc/install) or
on Debian or Ubuntu, you can install it from the package manager:

```bash
sudo apt-get install -y golang
```

To build, first ensure you have the protocol compiler (protoc) and markdown
processing library (blackfriday):

```bash
go get github.com/golang/protobuf/proto && go get -u gopkg.in/russross/blackfriday.v2
```
To build, run the following command from this project directory:

```bash
go build
```

Then ensure the resulting `protoc-gen-docs` binary is in your `PATH`. A recommended location
is `$HOME/bin`:

```bash
cp protoc-gen-docs $HOME/bin
```

Since the following is often in your `$HOME/.bashrc` file:

```bash
export PATH=$HOME/bin:$PATH
```

## Use protoc-gen-docs

Then to generate a page of HTML describing the protobuf defined by file.proto, run

```bash
protoc --docs_out=output_directory input_directory/file.proto
```


With that input, the output will be written to

	output_directory/file.pb.html

Using the `mode` option, you can control the output format from the plugin. The
html_page option is the default and produces a fully self-contained HTML page.
The html_fragment option outputs an HTML fragment that can be used to embed in a
larger page. Finally, the jekyll_html option outputs an HTML fragment augmented
with [Jekyll front-matter](https://jekyllrb.com/docs/frontmatter/)

You specify the mode option using this syntax:

```bash
protoc --docs_out=mode=html_page:output_directory input_directory/file.proto
```

Using the `warnings` option, you can control whether warnings are produced
to report proto elements that aren't commented. You can use this option with
the following syntax:

```bash
protoc --docs_out=warnings=true:output_directory input_directory/file.proto
```

You can specify both the mode and warnings options by separating them with commas:

```bash
protoc --docs_out=warnings=true,mode=html_page:output_directory input_directory/file.proto
```

# Writing Docs

Writing documentation for use with protoc-gen-docs is simply a matter of adding comments to elements
within the input proto file. You can put comments directly above individual elements, or to the
right. For example:

```proto
// A package-level comment
package pkg;

// This documents the message as a whole
message MyMsg {
    // This documents this field 
    // It can contain many lines.
    int32 field1 = 1;

    int32 field2 = 2;       // This documents field2
}
```

Comments are treated as markdown. You can thus embed classic markdown annotations within any comment.

## Linking to types and elements

In addition to normal markdown links, you can also use special proto links within any comment. Proto
links are used to create a link to other types or elements within the set of protos. You specify proto links
using two pairs of square brackets such as:

```proto

// This is a comment that links to another type: [MyOtherType][MyPkg.MyOtherType]
message MyMsg {

}

```

The first square brackets contain the name of the type to display in the resulting documentation. The second
square brackets contain the fully qualified name of the type or element being referenced, including the
package name.

## Annotations

Within a proto file, you can insert special comments which provide additional metadata to
use in producing quality documentation. Within a package, optionally include an unattached
comment of the form:

```
// $title: My Title
// $overview: My Overview
// $location: https://mysite.com/mypage.html
```

`$title` provides a title for the generated package documentation. This is used for things like the
title of the generated HTML. `$overview` is a one-line description of the package, useful for
tables of contents or indexes. Finally, `$location` indicates the expected URL for the generated
documentation. This is used to help downstream processing tools to know where to copy
the documentation, and is used when creating documentation links from other packages to this one.

You can also use the $front_matter annotation to introduce new Jekyll front matter when generating
Jekyll-friendly HTML. For example:

```
// $front_matter: order: 10
```

The above will include the front matter `order: 10` in the generated Jekyll HTML document.

If a comment for an element contains the annotation `$hide_from_docs`,
then the associated element will be omitted from the output. This is useful when staging the
introduction of new features that aren't quite ready for use yet. The annotation can appear
anywhere in the comment for the element. For example:

```proto
message MyMsg {
    int32 field1 = 1; // $hide_from_docs
}
```

The comment for any element can contain the annotation `$class: <foo>` which is used
to insert a specific HTML class around the generated element. This is useful to give
particular styling to particular elements. Common examples of useful classes include

```proto
message MyMsg {
    int32 field1 = 1; // $class: alpha
    int32 field2 = 2; // $class: beta
    int32 field3 = 3; // $class: experimental
}
```
