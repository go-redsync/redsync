module github.com/go-redsync/redsync/v2

go 1.13

require (
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/hashicorp/go-multierror v1.0.0
	github.com/stvp/tempredis v0.0.0-20181119212430-b82af8480203
)

// TODO: Remove this once this issue is addressed, or redigo no longer points to
//       v2.0.0+incompatible, above: https://github.com/gomodule/redigo/issues/366
replace github.com/gomodule/redigo => github.com/gomodule/redigo v1.7.0
