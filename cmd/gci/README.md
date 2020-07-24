# gci

`gci`, a tool that control golang package import order and make it always deterministic.

It is from [repo](https://github.com/daixiang0/gci) v0.1 release version.

## Examples

Run `gci -w -local github.com/daixiang0/gci main.go` and you will handle following cases.

### simple case

```go
package main
import (
  "golang.org/x/tools"

  "fmt"

  "github.com/daixiang0/gci"
)
```

to

```go
package main
import (
  "fmt"

  "golang.org/x/tools"

  "github.com/daixiang0/gci"
)
```

### with alias

```go
package main
import (
  "fmt"
  go "github.com/golang"
  "github.com/daixiang0"
)
```

to

```go
package main
import (
  "fmt"

  go "github.com/golang"

  "github.com/daixiang0/gci"
)
```

### with comment and alias

```go
package main
import (
  "fmt"
  _ "github.com/golang" // golang
  "github.com/daixiang0"
)
```

to

```go
package main
import (
  "fmt"

  // golang
  _ "github.com/golang"

  "github.com/daixiang0/gci"
)
```

