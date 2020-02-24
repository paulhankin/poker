Poker hand evaluator
====================

This go package provides a fast poker hand evaluator for 3-card,
5-card and 7-card hands.

Build modes
-----------

There are three modes of using this package, which can be chosen
with built tags. These affect how the data tables are constructed.

First is the default (generate) in which case a few seconds will be spent at
binary startup time generating lookup tables.

Second is "-tags staticdata" in which case a large (70MB) source file
is compiled into the package, which contains the data tables. This makes
the binary roughly 30MB bigger, and also compiles much slower.

Third is "-tags filedata" in which case the "poker.dat"
file must be in the current directory, and it's loaded at startup time.

Timings on my workstation to build and run "cmd/holdemeval", running with
arguments `./holdemeval -hands "AdAh QsQd 6c5c"`:

    * `-tags filedata`: 0.40sec to build, 0.29sec to run
    * `-tags staticdata`: 8.6sec to build, 0.267sec to run
    * default (gendata) : 0.42sec to build, 2.4sec to run

My machine has 12 CPU cores, so with fewer cores the startup cost of generating
the tables will be scaled up as you'd expect.

My recommendation is to use `-tags filedata` for development, `-tags staticdata` for
a release binary that is expected to run quickly, and the default if you don't
mind the slow startup time.

Documentation
-------------

Here's the [package documentation on godoc](https://godoc.org/github.com/paulhankin/poker)

The code is MIT licensed (see LICENSE.txt).
