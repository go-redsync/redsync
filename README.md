# Redsync

[![Build Status](https://travis-ci.org/go-redsync/redsync.svg?branch=master)](https://travis-ci.org/go-redsync/redsync)

Redsync provides a Redis-based distributed mutual exclusion lock implementation for Go as described in [this post](http://redis.io/topics/distlock). A reference library (by [antirez](https://github.com/antirez)) for Ruby is available at [github.com/antirez/redlock-rb](https://github.com/antirez/redlock-rb).

## Installation

Install Redsync using the go get command:

    $ go get gopkg.in/redsync.v1

Two driver implementations will be installed; however, only the one used will be include in your project.

 * [Redigo](https://github.com/gomodule/redigo).
 * [Go-redis](https://github.com/go-redis/redis).
 
See the [examples](examples) folder for usage of each driver.

## Documentation

- [Reference](https://godoc.org/gopkg.in/redsync.v1)

## Contributing

Contributions are welcome.

## License

Redsync is available under the [BSD (3-Clause) License](https://opensource.org/licenses/BSD-3-Clause).

## Disclaimer

This code implements an algorithm which is currently a proposal, it was not formally analyzed. Make sure to understand how it works before using it in production environments.
