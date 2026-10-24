# The CPAN modules the lab's Perl uses, for `cpanm --installdeps .`.
# Perl is optional in the lab. HOWTO.md, under Perl, says how to keep
# this file tidy: every module that perl/bin, perl/lib or perl/t uses
# directly, at the version it was tested with, and nothing else.

requires 'perl', '5.036';

requires 'Inline', '0.87';

# 0.58 needs a one-line fix on macOS (upstream issue #44). Install it as
# HOWTO.md describes first; cpanm then finds it already there.
requires 'Inline::Python', '0.58';
