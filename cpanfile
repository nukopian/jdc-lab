# The CPAN modules the lab's Perl uses, for
# `cpanm --installdeps --with-develop .` (leave out --with-develop to
# skip the tools in the develop block).
# Perl is optional in the lab. docs/MAINTAINING.md, under Perl, says how
# to keep this file tidy: every module that perl/bin, perl/lib or perl/t
# uses directly, at the version it was tested with, and nothing else.

requires 'perl', '5.036';

requires 'Inline', '0.87';

# Built against the Python that INLINE_PYTHON_EXECUTABLE names: the
# venv's, as docs/SETUP.md says.
requires 'Inline::Python', '0.59';

# The lab's test toolchain, Test2 rather than Test::More, listed ahead of
# the first test in perl/t. Test2::V0 comes with Test-Simple.
on test => sub {
    requires 'Test2::Plugin::NoWarnings', '0.10';
    requires 'Test2::Tools::Explain', '0.02';
    requires 'Test2::V0', '1.302226';
};

on develop => sub {
    # perlcritic, which reviews the code against .perlcriticrc.
    requires 'Perl::Critic', '1.156';

    # scan-perl-prereqs, for checking this file against the code.
    requires 'Perl::PrereqScanner', '1.100';

    # perltidy, which formats the code with .perltidyrc.
    requires 'Perl::Tidy', '20260826';

    # yath, which runs the tests, rather than prove.
    requires 'Test2::Harness', '1.000180';

    # Test2's UUID support, which yath uses, warns unless it finds UUID,
    # its preferred backend.
    requires 'UUID', '0.40';
};
