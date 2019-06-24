// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this currentFile except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/client9/gospell"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/russross/blackfriday"

	"istio.io/tools/pkg/protomodel"
)

type outputMode int

const (
	htmlPage                    outputMode = iota // stand-alone HTML page
	htmlFragment                                  // core portion of an HTML body, no head section or other wrappers
	htmlFragmentWithFrontMatter                   // like a fragment, but with YAML front-matter
)

type htmlGenerator struct {
	buffer           bytes.Buffer
	model            *protomodel.Model
	mode             outputMode
	numWarnings      int
	speller          *gospell.GoSpell
	customStyleSheet string

	// transient state as individual files are processed
	currentPackage             *protomodel.PackageDescriptor
	currentFrontMatterProvider *protomodel.FileDescriptor
	grouping                   bool

	genWarnings      bool
	warningsAsErrors bool
	emitYAML         bool
	camelCaseFields  bool
	perFile          bool
}

const (
	deprecated = "deprecated "
)

func newHTMLGenerator(model *protomodel.Model, mode outputMode, genWarnings bool, warningsAsErrors bool, speller *gospell.GoSpell,
	emitYAML bool, camelCaseFields bool, customStyleSheet string, perFile bool) *htmlGenerator {
	return &htmlGenerator{
		model:            model,
		mode:             mode,
		speller:          speller,
		genWarnings:      genWarnings,
		warningsAsErrors: warningsAsErrors,
		emitYAML:         emitYAML,
		camelCaseFields:  camelCaseFields,
		customStyleSheet: customStyleSheet,
		perFile:          perFile,
	}
}

func (g *htmlGenerator) getFileContents(file *protomodel.FileDescriptor,
	messages map[string]*protomodel.MessageDescriptor,
	enums map[string]*protomodel.EnumDescriptor,
	services map[string]*protomodel.ServiceDescriptor) {
	for _, m := range file.AllMessages {
		messages[g.relativeName(m)] = m
		g.includeUnsituatedDependencies(messages, enums, m)
	}

	for _, e := range file.AllEnums {
		enums[g.relativeName(e)] = e
	}

	for _, s := range file.Services {
		services[g.relativeName(s)] = s
	}
}

func (g *htmlGenerator) generatePerFileOutput(filesToGen map[*protomodel.FileDescriptor]bool, pkg *protomodel.PackageDescriptor,
	response *plugin.CodeGeneratorResponse) {
	// We need to produce a file for each non-hidden file in this package.

	// Decide which types need to be included in the generated file.
	// This will be all the types in the fileToGen input files, along with any
	// dependent types which are located in files that don't have
	// a known location on the web.

	for _, file := range pkg.Files {
		if _, ok := filesToGen[file]; ok {
			g.currentFrontMatterProvider = file
			messages := make(map[string]*protomodel.MessageDescriptor)
			enums := make(map[string]*protomodel.EnumDescriptor)
			services := make(map[string]*protomodel.ServiceDescriptor)

			g.getFileContents(file, messages, enums, services)
			var filename = path.Base(file.GetName())
			var extension = path.Ext(filename)
			var name = filename[0 : len(filename)-len(extension)]

			rf := g.generateFile(name, file, messages, enums, services)
			response.File = append(response.File, &rf)
		}
	}
}

func (g *htmlGenerator) generatePerPackageOutput(filesToGen map[*protomodel.FileDescriptor]bool, pkg *protomodel.PackageDescriptor,
	response *plugin.CodeGeneratorResponse) {
	// We need to produce a file for this package.

	// Decide which types need to be included in the generated file.
	// This will be all the types in the fileToGen input files, along with any
	// dependent types which are located in packages that don't have
	// a known location on the web.
	messages := make(map[string]*protomodel.MessageDescriptor)
	enums := make(map[string]*protomodel.EnumDescriptor)
	services := make(map[string]*protomodel.ServiceDescriptor)

	for _, file := range pkg.Files {
		if _, ok := filesToGen[file]; ok {
			g.getFileContents(file, messages, enums, services)
		}
	}

	rf := g.generateFile(pkg.Name, pkg.FileDesc(), messages, enums, services)
	response.File = append(response.File, &rf)
}

func (g *htmlGenerator) generateOutput(filesToGen map[*protomodel.FileDescriptor]bool) (*plugin.CodeGeneratorResponse, error) {
	// process each package; we produce one output file per package
	response := plugin.CodeGeneratorResponse{}

	for _, pkg := range g.model.Packages {
		g.currentPackage = pkg
		g.currentFrontMatterProvider = pkg.FileDesc()

		// anything to output for this package?
		count := 0
		for _, file := range pkg.Files {
			if _, ok := filesToGen[file]; ok {
				count++
			}
		}

		if count > 0 {
			if g.perFile {
				g.generatePerFileOutput(filesToGen, pkg, &response)
			} else {
				g.generatePerPackageOutput(filesToGen, pkg, &response)
			}
		}
	}

	if g.warningsAsErrors && g.numWarnings > 0 {
		return nil, fmt.Errorf("treating %d warnings as errors", g.numWarnings)
	}

	return &response, nil
}

func (g *htmlGenerator) descLocation(desc protomodel.CoreDesc) string {
	if g.perFile {
		return desc.FileDesc().Matter.HomeLocation
	}
	if desc.PackageDesc().FileDesc() != nil {
		return desc.PackageDesc().FileDesc().Matter.HomeLocation
	}
	return ""
}

func (g *htmlGenerator) includeUnsituatedDependencies(messages map[string]*protomodel.MessageDescriptor,
	enums map[string]*protomodel.EnumDescriptor,
	msg *protomodel.MessageDescriptor) {
	for _, field := range msg.Fields {
		switch f := field.FieldType.(type) {
		case *protomodel.MessageDescriptor:
			// A package without a known documentation location is included in the output.
			if g.descLocation(field.FieldType) == "" {
				name := g.relativeName(f)
				if _, ok := messages[name]; !ok {
					messages[name] = f
					g.includeUnsituatedDependencies(messages, enums, msg)
				}
			}
		case *protomodel.EnumDescriptor:
			if g.descLocation(field.FieldType) == "" {
				enums[g.relativeName(f)] = f
			}
		}
	}
}

// Generate a package documentation file or a collection of cross-linked files.
func (g *htmlGenerator) generateFile(name string, top *protomodel.FileDescriptor, messages map[string]*protomodel.MessageDescriptor,
	enums map[string]*protomodel.EnumDescriptor, services map[string]*protomodel.ServiceDescriptor) plugin.CodeGeneratorResponse_File {
	g.buffer.Reset()

	var typeList []string
	var serviceList []string

	for name, enum := range enums {
		if enum.IsHidden() {
			continue
		}

		absName := g.absoluteName(enum)
		known := wellKnownTypes[absName]
		if known != "" {
			continue
		}

		typeList = append(typeList, name)
	}

	for name, msg := range messages {
		// Don't generate virtual messages for maps.
		if msg.GetOptions().GetMapEntry() {
			continue
		}

		if msg.IsHidden() {
			continue
		}

		absName := g.absoluteName(msg)
		known := wellKnownTypes[absName]
		if known != "" {
			continue
		}

		typeList = append(typeList, name)
	}
	sort.Strings(typeList)

	for name, svc := range services {
		if svc.IsHidden() {
			continue
		}

		serviceList = append(serviceList, name)
	}
	sort.Strings(serviceList)

	numKinds := 0
	if len(typeList) > 0 {
		numKinds++
	}
	if len(serviceList) > 0 {
		numKinds++
	}

	// if there's more than one kind of thing, divide the output in groups
	g.grouping = numKinds > 1

	g.generateFileHeader(top, len(typeList)+len(serviceList))

	if len(serviceList) > 0 {
		if g.grouping {
			g.emit("<h2 id=\"Services\">Services</h2>")
		}

		for _, name := range serviceList {
			service := services[name]
			g.generateService(service)
		}
	}

	if len(typeList) > 0 {
		if g.grouping {
			g.emit("<h2 id=\"Types\">Types</h2>")
		}

		for _, name := range typeList {
			if e, ok := enums[name]; ok {
				g.generateEnum(e)
			} else if m, ok := messages[name]; ok {
				g.generateMessage(m)
			}
		}
	}

	g.generateFileFooter()

	return plugin.CodeGeneratorResponse_File{
		Name:    proto.String(name + ".pb.html"),
		Content: proto.String(g.buffer.String()),
	}
}

func (g *htmlGenerator) generateFileHeader(top *protomodel.FileDescriptor, numEntries int) {
	name := g.currentPackage.Name
	if g.mode == htmlFragmentWithFrontMatter {
		g.emit("---")

		if top != nil && top.Matter.Title != "" {
			g.emit("title: ", top.Matter.Title)
		} else {
			g.emit("title: ", name)
		}

		if top != nil && top.Matter.Overview != "" {
			g.emit("overview: ", top.Matter.Overview)
		}

		if top != nil && top.Matter.Description != "" {
			g.emit("description: ", top.Matter.Description)
		}

		if top != nil && top.Matter.HomeLocation != "" {
			g.emit("location: ", top.Matter.HomeLocation)
		}

		g.emit("layout: protoc-gen-docs")
		g.emit("generator: protoc-gen-docs")

		// emit additional custom front-matter fields
		if g.perFile {
			if top != nil {
				for _, fm := range top.Matter.Extra {
					g.emit(fm)
				}
			}
		} else {
			// Front matter may be in any of the package's files.
			for _, file := range g.currentPackage.Files {
				for _, fm := range file.Matter.Extra {
					g.emit(fm)
				}
			}
		}

		g.emit("number_of_entries: ", strconv.Itoa(numEntries))
		g.emit("---")
	} else if g.mode == htmlPage {
		g.emit("<!DOCTYPE html>")
		g.emit("<html itemscope itemtype=\"https://schema.org/WebPage\">")
		g.emit("<!-- Generated by protoc-gen-docs -->")
		g.emit("<head>")
		g.emit("<meta charset=\"utf-8'>")
		g.emit("<meta name=\"viewport' content=\"width=device-width, initial-scale=1, shrink-to-fit=no\">")

		if top != nil && top.Matter.Title != "" {
			g.emit("<meta name=\"title\" content=\"", top.Matter.Title, "\">")
			g.emit("<meta name=\"og:title\" content=\"", top.Matter.Title, "\">")
			g.emit("<title>", top.Matter.Title, "</title>")
		}

		if top != nil && top.Matter.Overview != "" {
			g.emit("<meta name=\"description\" content=\"", top.Matter.Overview, "\">")
			g.emit("<meta name=\"og:description\" content=\"", top.Matter.Overview, "\">")
		} else if top != nil && top.Matter.Description != "" {
			g.emit("<meta name=\"description\" content=\"", top.Matter.Description, "\">")
			g.emit("<meta name=\"og:description\" content=\"", top.Matter.Description, "\">")
		}

		if g.customStyleSheet != "" {
			g.emit("<link rel=\"stylesheet\" href=\"" + g.customStyleSheet + "\">")
		} else {
			g.emit(htmlStyle)
		}

		g.emit("</head>")
		g.emit("<body>")
		if top != nil && top.Matter.Title != "" {
			g.emit("<h1>", top.Matter.Title, "</h1>")
		}
	} else if g.mode == htmlFragment {
		g.emit("<!-- Generated by protoc-gen-docs -->")
		if top != nil && top.Matter.Title != "" {
			g.emit("<h1>", top.Matter.Title, "</h1>")
		}
	}

	if g.perFile {
		if top != nil {
			g.generateComment(top.Matter.Location, name)
		}
	} else {
		g.generateComment(g.currentPackage.Location(), name)
	}
}

func (g *htmlGenerator) generateFileFooter() {
	if g.mode == htmlPage {
		g.emit("</body>")
		g.emit("</html>")
	}
}

func (g *htmlGenerator) generateSectionHeading(desc protomodel.CoreDesc) {
	class := ""
	if desc.Class() != "" {
		class = desc.Class() + " "
	}

	heading := "h2"
	if g.grouping {
		heading = "h3"
	}

	name := g.relativeName(desc)

	g.emit("<", heading, " id=\"", normalizeID(name), "\">", name, "</", heading, ">")

	if class != "" {
		g.emit("<section class=\"", class, "\">")
	} else {
		g.emit("<section>")
	}
}

func (g *htmlGenerator) generateSectionTrailing() {
	g.emit("</section>")
}

func (g *htmlGenerator) generateMessage(message *protomodel.MessageDescriptor) {
	g.generateSectionHeading(message)
	g.generateComment(message.Location(), message.GetName())

	if len(message.Fields) > 0 {
		g.emit("<table class=\"message-fields\">")
		g.emit("<thead>")
		g.emit("<tr>")
		g.emit("<th>Field</th>")
		g.emit("<th>Type</th>")
		g.emit("<th>Description</th>")
		g.emit("</tr>")
		g.emit("</thead>")
		g.emit("<tbody>")

		var oneof int32 = -1
		for _, field := range message.Fields {
			if field.IsHidden() {
				continue
			}

			fieldName := *field.Name
			if g.camelCaseFields {
				fieldName = camelCase(*field.Name)
			}

			fieldTypeName := g.fieldTypeName(field)

			class := ""
			if field.Options != nil && field.Options.GetDeprecated() {
				class = deprecated
			}

			if field.Class() != "" {
				class = class + field.Class() + " "
			}

			if field.OneofIndex != nil {
				if *field.OneofIndex != oneof {
					class += "oneof oneof-start"
					oneof = *field.OneofIndex
				} else {
					class += "oneof"
				}
			}

			if class != "" {
				g.emit("<tr id=\"", normalizeID(g.relativeName(field)), "\" class=\"", class, "\">")
			} else {
				g.emit("<tr id=\"", normalizeID(g.relativeName(field)), "\">")
			}

			g.emit("<td><code>", fieldName, "</code></td>")
			g.emit("<td><code>", g.linkify(field.FieldType, fieldTypeName), "</code></td>")
			g.emit("<td>")

			g.generateComment(field.Location(), field.GetName())

			g.emit("</td>")
			g.emit("</tr>")
		}
		g.emit("</tbody>")
		g.emit("</table>")
	}

	g.generateSectionTrailing()
}

func (g *htmlGenerator) generateEnum(enum *protomodel.EnumDescriptor) {
	g.generateSectionHeading(enum)
	g.generateComment(enum.Location(), enum.GetName())

	if len(enum.Values) > 0 {
		g.emit("<table class=\"enum-values\">")
		g.emit("<thead>")
		g.emit("<tr>")
		g.emit("<th>Name</th>")
		g.emit("<th>Description</th>")
		g.emit("</tr>")
		g.emit("</thead>")
		g.emit("<tbody>")

		for _, v := range enum.Values {
			if v.IsHidden() {
				continue
			}

			name := *v.Name

			class := ""
			if v.Options != nil && v.Options.GetDeprecated() {
				class = deprecated
			}

			if v.Class() != "" {
				class = class + v.Class() + " "
			}

			if class != "" {
				g.emit("<tr id=\"", normalizeID(g.relativeName(v)), "\" class=\"", class, "\">")
			} else {
				g.emit("<tr id=\"", normalizeID(g.relativeName(v)), "\">")
			}
			g.emit("<td><code>", name, "</code></td>")
			g.emit("<td>")

			g.generateComment(v.Location(), name)

			g.emit("</td>")
			g.emit("</tr>")
		}
		g.emit("</tbody>")
		g.emit("</table>")
	}

	g.generateSectionTrailing()
}

func (g *htmlGenerator) generateService(service *protomodel.ServiceDescriptor) {
	g.generateSectionHeading(service)
	g.generateComment(service.Location(), service.GetName())

	for _, method := range service.Methods {
		if method.IsHidden() {
			continue
		}

		class := ""
		if method.Options != nil && method.Options.GetDeprecated() {
			class = deprecated
		}

		if method.Class() != "" {
			class = class + method.Class() + " "
		}

		if class != "" {
			g.emit("<pre id=\"", normalizeID(g.relativeName(method)), "\" class=\"", class, "\"><code class=\"language-proto\">rpc ",
				method.GetName(), "(", g.relativeName(method.Input), ") returns (", g.relativeName(method.Output), ")")
		} else {
			g.emit("<pre id=\"", normalizeID(g.relativeName(method)), "\"><code class=\"language-proto\">rpc ",
				method.GetName(), "(", g.relativeName(method.Input), ") returns (", g.relativeName(method.Output), ")")
		}
		g.emit("</code></pre>")

		g.generateComment(method.Location(), method.GetName())
	}

	g.generateSectionTrailing()
}

// emit prints the arguments to the generated output.
func (g *htmlGenerator) emit(str ...string) {
	for _, s := range str {
		g.buffer.WriteString(s)
	}
	g.buffer.WriteByte('\n')
}

var typeLinkPattern = regexp.MustCompile(`\[[^]]*]\[[^]]*]`)

func (g *htmlGenerator) generateComment(loc protomodel.LocationDescriptor, name string) {
	com := loc.GetLeadingComments()
	if com == "" {
		com = loc.GetTrailingComments()
		if com == "" {
			g.warn(loc, 0, "no comment found for %s", name)
			return
		}
	}

	text := strings.TrimSuffix(com, "\n")
	lines := strings.Split(text, "\n")
	if len(lines) > 0 {
		// Based on the amount of spacing at the start of the first line,
		// remove that many characters at the beginning of every line in the comment.
		// This is so we don't inject extra spaces in any preformatted blocks included
		// in the comments
		pad := 0
		for i, ch := range lines[0] {
			if !unicode.IsSpace(ch) {
				pad = i
				break
			}
		}

		for i := 0; i < len(lines); i++ {
			l := lines[i]
			if len(l) > pad {
				skip := 0
				var ch rune
				for skip, ch = range l {
					if !unicode.IsSpace(ch) {
						break
					}

					if skip == pad {
						break
					}
				}
				lines[i] = l[skip:]
			}
		}

		// now, adjust any headers included in the comment to correspond to the right
		// level, based on the heading level of the surrounding content
		for i := 0; i < len(lines); i++ {
			l := lines[i]
			if strings.HasPrefix(l, "#") {
				if g.grouping {
					lines[i] = "###" + l
				} else {
					lines[i] = "##" + l
				}
			}
		}

		// Replace any < and > in the text with HTML entities to avoid bad HTML output
		for i := 0; i < len(lines); i++ {
			lines[i] = strings.Replace(lines[i], "<", "&lt;", -1)
			lines[i] = strings.Replace(lines[i], ">", "&gt;", -1)
		}

		// find any type links of the form [name][type] and turn
		// them into normal HTML links
		for i := 0; i < len(lines); i++ {
			lines[i] = typeLinkPattern.ReplaceAllStringFunc(lines[i], func(match string) string {
				end := 0
				for match[end] != ']' {
					end++
				}

				linkName := match[1:end]
				typeName := match[end+2 : len(match)-1]

				if o, ok := g.model.AllDescByName["."+typeName]; ok {
					return g.linkify(o, linkName)
				}

				if l, ok := wellKnownTypes[typeName]; ok {
					return "<a href=\"" + l + "\">" + linkName + "</a>"
				}

				g.warn(loc, -(len(lines) - i), "unresolved type link [%s][%s]", linkName, typeName)

				return "*" + linkName + "*"
			})
		}
	}

	text = strings.Join(lines, "\n")

	if g.speller != nil {
		preBlock := false
		for linenum, line := range lines {
			trimmed := strings.Trim(line, " ")
			if strings.HasPrefix(trimmed, "```") {
				preBlock = !preBlock
				continue
			}

			if preBlock {
				continue
			}

			line := sanitize(line)

			words := g.speller.Split(line)
			for _, word := range words {
				if !g.speller.Spell(word) {
					g.warn(loc, -(len(lines) - linenum), "%s is misspelled", word)
				}
			}
		}
	}

	// turn the comment from markdown into HTML
	result := blackfriday.Run([]byte(text), blackfriday.WithExtensions(blackfriday.FencedCode|blackfriday.AutoHeadingIDs))

	// compensate for an unexpected Blackfriday bug, where it incorrect converts expands the & in HTML entites to &amp;
	result = bytes.Replace(result, []byte("&amp;lt;"), []byte("&lt;"), -1)
	result = bytes.Replace(result, []byte("&amp;gt;"), []byte("&gt;"), -1)

	// prevent any { contained in the markdown from being interpreted as Hugo shortcodes
	result = bytes.Replace(result, []byte("{"), []byte("&lbrace;"), -1)

	g.buffer.Write(result)
	g.buffer.WriteByte('\n')
}

var stripCodeBlocks = regexp.MustCompile("(`.*`)")
var stripMarkdownURLs = regexp.MustCompile(`\[.*\]\((.*)\)`)
var stripHTMLURLs = regexp.MustCompile(`(<a href=".*">)`)

func sanitize(line string) string {
	// strip out any embedded code blocks and URLs
	line = stripMarkdownURLs.ReplaceAllString(line, "")
	line = stripHTMLURLs.ReplaceAllString(line, "")
	line = stripCodeBlocks.ReplaceAllString(line, "")
	return line
}

// well-known types whose documentation we can refer to
var wellKnownTypes = map[string]string{
	"google.protobuf.Duration":    "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#duration",
	"google.protobuf.Timestamp":   "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#timestamp",
	"google.protobuf.Any":         "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#any",
	"google.protobuf.BytesValue":  "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#bytesvalue",
	"google.protobuf.StringValue": "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#stringvalue",
	"google.protobuf.BoolValue":   "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#boolvalue",
	"google.protobuf.Int32Value":  "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#int32value",
	"google.protobuf.Int64Value":  "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#int64value",
	"google.protobuf.Uint32Value": "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#uint32value",
	"google.protobuf.Uint64Value": "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#uint64value",
	"google.protobuf.FloatValue":  "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#floatvalue",
	"google.protobuf.DoubleValue": "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#doublevalue",
	"google.protobuf.Empty":       "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#empty",
	"google.protobuf.EnumValue":   "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#enumvalue",
	"google.protobuf.ListValue":   "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#listvalue",
	"google.protobuf.NullValue":   "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#nullvalue",
	"google.protobuf.Struct":      "https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#struct",
}

func (g *htmlGenerator) linkify(o protomodel.CoreDesc, name string) string {
	if o == nil {
		return name
	}

	if msg, ok := o.(*protomodel.MessageDescriptor); ok && msg.GetOptions().GetMapEntry() {
		return name
	}

	known := wellKnownTypes[g.absoluteName(o)]
	if known != "" {
		return "<a href=\"" + known + "\">" + name + "</a>"
	}

	if !o.IsHidden() {
		// is there a file-specific home location?
		loc := o.FileDesc().Matter.HomeLocation

		// if there isn't a file-specific home location, see if there's a package-wide location
		if loc == "" && o.PackageDesc().FileDesc() != nil {
			loc = o.PackageDesc().FileDesc().Matter.HomeLocation
		}

		if loc != "" && (g.currentFrontMatterProvider == nil || loc != g.currentFrontMatterProvider.Matter.HomeLocation) {
			return "<a href=\"" + loc + "#" + normalizeID(protomodel.DottedName(o)) + "\">" + name + "</a>"
		}
	}

	return "<a href=\"#" + normalizeID(g.relativeName(o)) + "\">" + name + "</a>"
}

func (g *htmlGenerator) warn(loc protomodel.LocationDescriptor, lineOffset int, format string, args ...interface{}) {
	if g.genWarnings {
		place := ""
		if loc.SourceCodeInfo_Location != nil && len(loc.Span) >= 2 {

			if lineOffset < 0 {
				place = fmt.Sprintf("%s:%d: ", loc.File.GetName(), loc.Span[0]+int32(lineOffset)+1)
			} else {
				place = fmt.Sprintf("%s:%d:%d: ", loc.File.GetName(), loc.Span[0]+1, loc.Span[1]+1)
			}
		}

		_, _ = fmt.Fprintf(os.Stderr, place+format+"\n", args...)
		g.numWarnings++
	}
}

func (g *htmlGenerator) relativeName(desc protomodel.CoreDesc) string {
	typeName := protomodel.DottedName(desc)
	if desc.PackageDesc() == g.currentPackage {
		return typeName
	}

	return desc.PackageDesc().Name + "." + typeName
}

func (g *htmlGenerator) absoluteName(desc protomodel.CoreDesc) string {
	typeName := protomodel.DottedName(desc)
	return desc.PackageDesc().Name + "." + typeName
}

func (g *htmlGenerator) fieldTypeName(field *protomodel.FieldDescriptor) string {
	name := "n/a"
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		name = "double"

	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		name = "float"

	case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		name = "int32"

	case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		name = "int64"

	case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
		name = "uint64"

	case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_FIXED32:
		name = "uint32"

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		name = "bool"

	case descriptor.FieldDescriptorProto_TYPE_STRING:
		name = "string"

	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		msg := field.FieldType.(*protomodel.MessageDescriptor)
		if msg.GetOptions().GetMapEntry() {
			keyType := g.fieldTypeName(msg.Fields[0])
			valType := g.linkify(msg.Fields[1].FieldType, g.fieldTypeName(msg.Fields[1]))
			return "map&lt;" + keyType + ",&nbsp;" + valType + "&gt;"
		}
		name = g.relativeName(field.FieldType)

	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		name = "bytes"

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		name = g.relativeName(field.FieldType)
	}

	if field.IsRepeated() {
		name += "[]"
	}

	if field.OneofIndex != nil {
		name += " (oneof)"
	}

	return name
}

/* TODO
func (g *htmlGenerator) fieldYAMLTypeName(field *FieldDescriptor) string {
	name := "n/a"
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		name = "double"

	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		name = "float"

	case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		name = "int32"

	case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		name = "int64"

	case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
		name = "uint64"

	case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_FIXED32:
		name = "uint32"

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		name = "bool"

	case descriptor.FieldDescriptorProto_TYPE_STRING:
		name = "string"

	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		msg := field.typ.(*MessageDescriptor)
		if msg.GetOptions().GetMapEntry() {
			keyType := g.fieldTypeName(msg.fields[0])
			valType := g.linkify(msg.fields[1].typ, g.fieldTypeName(msg.fields[1]))
			return "map&lt;" + keyType + ",&nbsp;" + valType + "&gt;"
		}
		name = g.relativeName(field.typ)

	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		name = "bytes"

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		name = "enum(" + g.relativeName(field.typ) + ")"
	}

	return name
}
*/

// camelCase returns the camelCased name.
func camelCase(s string) string {
	b := bytes.Buffer{}
	nextUpper := false
	for _, ch := range s {
		if ch == '_' {
			nextUpper = true
		} else {
			if nextUpper {
				nextUpper = false
				ch = unicode.ToUpper(ch)
			}
			b.WriteRune(ch)
		}
	}

	return b.String()
}

func normalizeID(id string) string {
	id = strings.Replace(id, " ", "-", -1)
	return strings.Replace(id, ".", "-", -1)
}

var htmlStyle = `
<style>
    html {
        overflow-y: scroll;
        position: relative;
        min-height: 100%
    }

    body {
        font-family: "Roboto", "Helvetica Neue", Helvetica, Arial, sans-serif;
        color: #535f61
    }

    a {
        color: #466BB0;
        text-decoration: none;
        font-weight: 500
    }

    a:hover, a:focus {
        color: #8ba3d1;
        text-decoration: none;
        font-weight: 500
    }

    a.disabled {
        color: #ccc;
        text-decoration: none;
        font-weight: 500
    }

    table, th, td {
        border: 1px solid #849396;
        padding: .3em
    }

	tr.oneof>td {
		border-bottom: 1px dashed #849396;
		border-top: 1px dashed #849396;
	}

    table {
        border-collapse: collapse
    }

    th {
        color: #fff;
        background-color: #286AC7;
        font-weight: normal
    }

    p {
        font-size: 1rem;
        line-height: 1.5;
        margin: .25em 0
    }

	table p:first-of-type {
		margin-top: 0
	}

	table p:last-of-type {
		margin-bottom: 0
	}

    @media (min-width: 768px) {
        p {
            margin: 1.5em 0
        }
    }

    li, dt, dd {
        font-size: 1rem;
        line-height: 1.5;
        margin: .25em
    }

    ol, ul, dl {
        list-style: initial;
        font-size: 1rem;
        margin: 0 1.5em;
        padding: 0
    }

    li p, dt p, dd p {
        margin: .4em 0
    }

    ol {
        list-style: decimal
    }

    h1, h2, h3, h4, h5, h6 {
        border: 0;
        font-weight: normal
    }

    h1 {
        font-size: 2.5rem;
        color: #286AC7;
        margin: 30px 0
    }

    h2 {
        font-size: 2rem;
        color: #2E2E2E;
        margin-bottom: 20px;
        margin-top: 30px;
        padding-bottom: 10px;
        border-bottom: 1px;
        border-color: #737373;
        border-style: solid
    }

    h3 {
        font-size: 1.85rem;
        font-weight: 500;
        color: #404040;
        letter-spacing: 1px;
        margin-bottom: 20px;
        margin-top: 30px
    }

    h4 {
        font-size: 1.85rem;
        font-weight: 500;
        margin: 30px 0 20px;
        color: #404040
    }

    em {
        font-style: italic
    }

    strong {
        font-weight: bold
    }

    blockquote {
        display: block;
        margin: 1em 3em;
        background-color: #f8f8f8
    }

	section {
		padding-left: 2em;
	}

	code {
		color: red;
	}

	.deprecated {
		background: silver;
	}

	.experimental {
		background: yellow;
	}
</style>
`
