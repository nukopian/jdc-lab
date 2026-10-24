use v5.36;

use Test2::V0;
use Test2::Plugin::NoWarnings;

# Perl is optional in the lab, so skip, rather than fail, without
# Inline::Python.
BEGIN {
    eval { require Inline::Python; 1 }
        or skip_all('no Inline::Python: see docs/SETUP.md');
}

use JDCLab::OnnxDemo qw(vectors similarities runtime_version);

# The scores for each pair of vectors, to six places: the same numbers the
# Python and Go tests check.
my %expected  = ('0,1' => 0.962805, '0,2' => 0.130990, '1,2' => 0.288312);
my $tolerance = 1e-6;

my @vectors = vectors();
my $scores  = similarities(\@vectors);

like(
    runtime_version(), qr/^\d+\.\d+\.\d+$/,
    'ONNX Runtime reports its version',
);

is(scalar @$scores, scalar @vectors, 'one row of scores for each vector');

for my $i (keys @vectors) {
    is(
        $scores->[$i][$i], float(1, tolerance => $tolerance),
        "vector $i is identical to itself",
    );
    for my $j (keys @vectors) {
        is(
            $scores->[$i][$j], float($scores->[$j][$i], tolerance => $tolerance),
            "score [$i][$j] equals [$j][$i]",
        );
    }
}

for my $pair (sort keys %expected) {
    my($i, $j) = split /,/, $pair;
    is(
        $scores->[$i][$j], float($expected{$pair}, tolerance => $tolerance),
        "score [$i][$j] is $expected{$pair}",
    );
}

done_testing;
