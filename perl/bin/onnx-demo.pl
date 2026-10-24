#!/usr/bin/env perl
# Runs the lab's small ONNX model from Perl, through Inline::Python and ONNX
# Runtime, and prints how alike each pair of its vectors is. The Python and
# Go demos print the same table. The work is in JDCLab::OnnxDemo (perl/lib),
# which perl/t/onnx-demo.t tests.
#
# From the lab's folder, with direnv on (see docs/SETUP.md):
#
#   perl perl/bin/onnx-demo.pl
use v5.36;

use JDCLab::OnnxDemo qw(names vectors similarities runtime_version);

my @names   = names();
my @vectors = vectors();

printf "ONNX Runtime %s, called from Perl %vd\n\n", runtime_version(), $^V;
my $scores = similarities(\@vectors);
printf "%-18s%s\n", '', join '', map { sprintf '%18s', $_ } @names;
for my $i (keys @names) {
    printf "%-18s%s\n", $names[$i], join '',
        map { sprintf '%18.3f', $_ } $scores->[$i]->@*;
}
