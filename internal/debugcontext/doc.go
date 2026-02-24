// Package debugcontext provides a passive, append-only data structure that
// accumulates diagnostic entries (errors, warnings, info, traces) as the
// assembler pipeline progresses. It does not perform I/O or formatting —
// a separate renderer consumes the entries to produce output.
//
// See .requirements/assembler/debug-information.md for the full specification.
package debugcontext
