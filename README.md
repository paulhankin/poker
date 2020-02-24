Poker hand evaluator
====================

This go package provides a fast poker hand evaluator for 3-card,
5-card and 7-card hands.

When benchmarking on my machine, on a single core I get around
36 million 5-card evaluations per second, or roughly 100 CPU cycles
per eval. And 25 million 7-card evaluations per second, or
roughly 150 cycles per eval.

It works using the same principles as the [2+2 hand evaluator](http://archives1.twoplustwo.com/showflat.php?Cat=0&Number=8513906).
It uses a huge state machine, with 52 transitions out for each state representing
52 possible next cards. The final nodes contain 52 ranks for each of the
52 last cards. By merging nodes for equivalent hands, the number of
states is much smaller than (52\*51\*50\*49\*48\*47\*46) as it would be
for a naive 7-card state machine. The number of states for the 5-card
eval is only 3459, and 163060 for the 7-card eval.

The novelty (or at least, I think it's novel) is that each transition
includes a remapping of suits to be applied to future cards, which greatly
reduces the number of states. (It could also be used to limit the number
of transitions for later states, where there are only effectively 2 suits:
the suit where a flush is possible, and the other suits. Clever use
of this could further reduce the size of the state table. This isn't done yet).

NOTE! You probably want to build any program that depends on
this package with "-tags staticdata" for releases to avoid a several-second
startup time. See the section on build modes.

TODO: further state merging as described above, and rewrite the eval code
in assembler, to avoid bounds checking. I guess the suit transforms can
be written faster.

Build modes
-----------

There are three modes of using this package, which can be chosen
with built tags. These affect how the data tables are constructed.

First is the default (generate) in which case a few seconds will be spent at
binary startup time generating lookup tables.

Second is "-tags staticdata" in which case a large (40MB) source file
is compiled into the package, which contains the data tables. This makes
the binary roughly 8MB bigger, but also compiles much slower. There is a
little bit of startup time (0.2s), because the tables are stored compressed
and are uncompressed at runtime.

Third is "-tags filedata" in which case the "poker.dat"
file must be in the current directory, and it's loaded at startup time.

Timings on my workstation to build and run "cmd/holdemeval", running with
arguments `./holdemeval -hands "AdAh QsQd 6c5c"`:

    * `-tags filedata`: 0.40sec to build, 0.29sec to run
    * `-tags staticdata`: 6.9sec to build, 0.44sec to run
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
