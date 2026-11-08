# mvp

An opinionated Go web application stack: routing, forms, templates, configuration, EDB storage, background jobs and live updates. The root package assembles these systems; it is a larger dependency than a standalone router.

## Install

Requires Go 1.23 or newer.

```sh
go get github.com/andreyvit/mvp@latest
```

## Example

```go
package main

import (
	"fmt"

	"github.com/andreyvit/mvp/backoff"
)

func main() {
	policy := backoff.GoodBackoff
	fmt.Println(policy.DelayAfter(1)) // 0s
	fmt.Println(policy.DelayAfter(2)) // 1s
}
```

The example uses the standalone `backoff` package; it does not start a web app. Stop retrying when `DelayAfter` returns `backoff.InfiniteDelay`.

For the full stack, start with [Configuration](settings-configuration.go), [Main](mvpmain.go), [App.Initialize](app.go) and [route registration](routing-builder.go). Application setup requires settings, hooks and storage configuration. Review those requirements before adopting the root package.

## License

See [LICENSE](LICENSE).
