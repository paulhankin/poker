Poker hand evaluator
====================

This go package provides a fast poker hand evaluator for 3-card,
5-card and 7-card hands.

Build modes
-----------

There are three modes of using this package, which can be chosen
with built tags. These affect how the data tables are constructed.

First is "-tags gendata" in which case a few seconds will be spent at
binary startup time generating lookup tables.

Second is "-tags staticdata" in which case a large (70MB) source file
is compiled into the package, which contains the data tables. This makes
the binary roughly 30MB bigger, and also compiles much slower.

Third is to not specify any build tags, in which case the "poker.dat"
file must be in the current directory, and it's loaded at startup time.

Documentation
-------------

Here's the [package documentation on godoc](https://godoc.org/github.com/paulhankin/poker)

The code is MIT licensed (see LICENSE.txt).
