# ADR 002: Spatial Margin & Watermark Scrubbing Heuristics

## Status
Accepted

## Context
PDF documents frequently contain invasive artifacts that disrupt Markdown readability and LLM token processing:
1. Running headers and footers containing page numbers, document revisions, and print timestamps.
2. Diagonal or full-page watermark stamps (e.g., "CONFIDENTIAL", "DRAFT", "INTERNAL USE ONLY").
3. Tracking codes, system UUIDs, or document hashes.

## Decision
We implemented a two-stage sanitization pipeline in `internal/cleaner`:
1. **Spatial Margin Coordinate Filtering**: Using standard 72-dpi coordinate thresholds (top margin $Y > 750\text{pt}$, bottom margin $Y < 55\text{pt}$), text matching pagination patterns (`Page \d+ of \d+`, `\d+/\d+`, single/double digits) is filtered.
2. **Lexical Watermark Matching**: Case-insensitive substring matching against a curated lexicon of confidential and draft watermarks.
3. **Configurable Regex Redaction**: User-supplied regular expressions are executed across all text spans to redact specific domain tokens before formatting.

## Consequences

### Positive
- Removes repetitive noise across multi-page documents without breaking body text continuity.
- Drastically improves token quality when feeding generated Markdown to LLMs or vector search databases.
- Configurable: can be selectively toggled on/off via `--clean` or adjusted via `--strip`.

### Negative
- Very short documents with non-standard page heights could potentially have valid text in margin regions if not designed using standard letter/A4 conventions. The `--clean=false` flag allows users to disable this when converting non-standard PDF formats.
