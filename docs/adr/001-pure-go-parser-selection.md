# ADR 001: Pure Go Architecture vs Cgo/Poppler

## Status
Accepted

## Context
A primary requirement for `pdf2md` is high performance, instant startup, and seamless cross-platform deployment on Windows and Linux without complex system installation prerequisites. Many existing PDF extraction tools rely on native C/C++ libraries such as Poppler (`pdftotext`), MuPDF, or Tesseract OCR, which require Cgo, dynamic library linking (`.dll` on Windows, `.so` on Linux), and complex C/C++ toolchains.

## Decision
We chose a **100% Pure Go architecture** utilizing `github.com/ledongthuc/pdf` for positional text extraction and `github.com/pdfcpu/pdfcpu` for asset management.

## Consequences

### Positive
- **Zero-Dependency Binaries**: Produces single, standalone static executables without requiring external C libraries or runtime DLLs.
- **Cross-Compilation**: Cross-compiling for Windows (`windows/amd64`) and Linux (`linux/amd64`) works natively via `GOOS` and `GOARCH` without cross-compilation toolchains.
- **High Concurrency**: Goroutines handle batch page and document parsing in parallel with zero memory safety hazards or Cgo bridge overhead.
- **Portability**: Enables frictionless CI/CD pipelines and straightforward installation on user machines.

### Negative
- Pure Go PDF parsers do not provide optical character recognition (OCR) for scanned, non-searchable image-only PDFs. For standard vector and digital PDFs, pure Go parsing is orders of magnitude faster and significantly simpler.
