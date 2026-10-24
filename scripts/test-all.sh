#!/bin/sh
# Runs the lab's tests in every language that's set up here: Python's with
# pytest, then ty's type check, Go's with go test, and Perl's with yath. A language that isn't set
# up is skipped, with a note saying so; any suite that fails fails the run.
# Run it from anywhere in the lab, with direnv on (see docs/SETUP.md):
#
#   test-all.sh
set -u

root=$(cd "$(dirname "$0")/.." && pwd)
cd "$root" || exit 2

# Put the lab's own tools first, as .envrc means them to be, in case a
# shell hook (mise's, for one) has moved them behind the system's since.
if [ -n "${PERLBREW_ROOT:-}" ] && [ -n "${PERLBREW_PERL:-}" ]; then
	PATH=$PERLBREW_ROOT/perls/$PERLBREW_PERL/bin:$PATH
	if [ -n "${PERLBREW_LIB:-}" ]; then
		PATH=${PERLBREW_HOME:-$HOME/.perlbrew}/libs/$PERLBREW_PERL@$PERLBREW_LIB/bin:$PATH
	fi
fi
if [ -d .venv/bin ]; then
	PATH=$root/.venv/bin:$PATH
fi
export PATH

failed=
summary=

# note LANGUAGE RESULT: remember a language's result for the summary.
note() {
	summary="$summary$(printf '%-8s %s' "$1" "$2")
"
	case $2 in FAIL*) failed=1 ;; esac
}

echo "== Python"
if [ -x .venv/bin/pytest ]; then
	# Type errors fail the run as test failures do. ty needs --project to
	# find its settings in python/pyproject.toml from the root.
	if .venv/bin/pytest python && .venv/bin/ty check --project python; then
		note Python pass
	else
		note Python FAIL
	fi
else
	note Python "skipped: no .venv"
fi

echo "== Go"
if command -v go >/dev/null 2>&1; then
	# The Go tests run the model the Python demo saves; make it if need be.
	if [ ! -f models/onnx-demo.onnx ] && [ -x .venv/bin/python ]; then
		PYTHONPATH=python/lib .venv/bin/python python/bin/onnx_demo.py >/dev/null
	fi
	if go test ./...; then note Go pass; else note Go FAIL; fi
else
	note Go "skipped: no go"
fi

echo "== Perl"
if ! command -v yath >/dev/null 2>&1; then
	note Perl "skipped: no yath (Perl is optional)"
elif ! perl -MInline::Python -e1 2>/dev/null; then
	note Perl "skipped: no Inline::Python (Perl is optional)"
else
	if yath test perl/t; then note Perl pass; else note Perl FAIL; fi
fi

echo "== Summary"
printf '%s' "$summary"
[ -z "$failed" ]
