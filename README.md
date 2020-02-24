Poker hand evaluator
====================

This go package provides a fast poker hand evaluator for 3-card,
5-card and 7-card hands.

There are two modes of using this package: one is to use the "gendata"
build tag, in which case a few seconds will be spent at binary startup
time generating lookup tables. The other is to not specify this tag,
which includes the tables (roughly 70MB) in the binary you build. This
has the side-effect of making your compiles slow, since the static tables
are included in a 80MB source file.

Here's the [package documentation on godoc](https://godoc.org/github.com/paulhankin/poker)

The code is MIT licensed (see LICENSE.txt).
