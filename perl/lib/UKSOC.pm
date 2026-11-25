use v5.36;

package UKSOC;

use Ekorn::Pipeline::Operators;

use Ekorn::Pipeline::Tooling qw(:all);
use Path::Tiny               qw(path);

# Found from this module, perl/lib/UKSOC.pm, so any script can use it.
our $DATA_DIR = path(__FILE__)->realpath->parent(3)->child('local/uksoc');
our $ABBREVIATIONS_TSV = path($DATA_DIR, 'abbreviations.tsv');
our $ENTRIES_TSV       = path($DATA_DIR, 'entries.tsv');
our $FRAMEWORK_TSV     = path($DATA_DIR, 'framework.tsv');
our $ISCO_TSV          = path($DATA_DIR, 'isco.tsv');
our $UNITS_TSV         = path($DATA_DIR, 'units.tsv');
our $MAJOR             = 1;
our $SUB_MAJOR         = 2;
our $MINOR             = 3;
our $UNIT              = 4;

unless (caller) {
    my %x = (
        Abbreviations => all_abbreviations(),
        Entries       => all_entries(),
        Framework     => all_frameworks(),
        ISCO          => all_isco(),
        Groups        => groups(),
        Units         => units(),
    );
    dump_to->($x{Groups});
}


sub level_as_string ($l) {
    state $mapping = {
        $MAJOR     => 'major',
        $SUB_MAJOR => 'sub-major',
        $MINOR     => 'minor',
        $UNIT      => 'unit',
    };

    return $mapping->{$l} // 'Level(' . (0+ $l) . ')';
}


sub all_abbreviations () {
    state $data = abbreviations() | keep_all;
    return $data;
}


sub abbreviations (@stages) {
    return pipeline(\&_as_abbreviation, @stages)
        ->iterator(tsv($ABBREVIATIONS_TSV));
}


sub _as_abbreviation {
    return {
        Abbreviation => $_->[0],
        Meaning      => $_->[1],
        Frequent     => $_->[2] eq '*',
    };
}


sub all_entries () {
    state $data = entries() | keep_all;
    return $data;
}


sub entries (@stages) {
    return pipeline(\&_as_entry, @stages)->iterator(tsv($ENTRIES_TSV));
}


sub _as_entry {
    return {
        ID         => $_->[0],
        RecNo      => $_->[1],
        Default    => $_->[2],
        SOC90      => $_->[3],
        SOC2000    => $_->[4],
        SOC2010    => $_->[5],
        SOC2020    => $_->[6],
        SOC2020Ext => $_->[7],
        Indexed    => $_->[8],
        Title      => $_->[9],
        Add        => $_->[10],
        Industry   => $_->[11],
        See        => $_->[12],
        ISCO08     => $_->[13],
        Version    => $_->[14],
        Change     => $_->[15],
    };
}


sub all_frameworks () {
    state $data = frameworks() | keep_all;
    return $data;
}


sub frameworks (@stages) {
    return pipeline(\&_as_framework, @stages)->iterator(tsv($FRAMEWORK_TSV));
}


sub _as_framework {
    return {
        ID      => $_->[0],
        Code    => $_->[1],
        Level   => length $_->[1],
        Title   => $_->[2],
        Change  => $_->[3],
        Version => $_->[4],
    };
}


sub all_isco () {
    state $data = isco() | keep_all;
    return $data;
}


sub isco (@stages) {
    return pipeline(\&_as_isco, @stages)->iterator(tsv($ISCO_TSV));
}


sub _as_isco {
    return {
        Code  => $_->[0],
        Level => length $_->[0],
        Title => $_->[1],
    };
}


# Each scheme's groups, by code. The sink takes [$scheme, $group] pairs,
# as units' takes [$scheme, $code, $title] rows, since the groups come
# from two sources.
sub groups () {
    state $groups = pipeline(
        groups_index_sink()
    )->iterator(from_array(
            (map { [SOC2020 => $_] } @{ all_frameworks() }),
            (map { [ISCO08  => $_] } @{ all_isco() }),
    ))->();
    return $groups;
}


sub groups_index_sink () {
    return sink_one {
        my %groups;
        return (
            sub {
                my($scheme, $group) = @$_;
                $groups{$scheme}{$group->{Code}} = $group;
            },
            sub {
                \%groups;
            },
        );
    };
}


sub units () {
    state $units = pipeline(
        units_index_sink()
    )->iterator(tsv($UNITS_TSV))->();
    return $units;
}


sub units_index_sink () {
    return sink_one {
        my %units;
        return (
            sub {
                my($scheme, $code, $title) = @$_;
                $units{$scheme}{$code} = $title;
            },
            sub {
                \%units;
            },
        );
    };
}


# A source of the rows of a TSV file, as arrays, its header skipped.
sub tsv ($filename, %options) {
    return from_csv(
        file   => $filename,
        header => q(skip),
        csv    => {sep_char => "\t"},
        %options,
    );
}

1;
