# `dodeval` — Definition of Done & Acceptance Criteria Evaluator

`dodeval` adalah *micro-package* mandiri (zero external dependencies) untuk mem-parsing checklist Definition of Done (DoD) atau kriteria penerimaan dari Markdown/teks, dan mengevaluasi kesesuaian bukti verifikasi laporan (*evidence gap analysis*).

## Fitur Utama

- **Zero External Dependencies**: Menggunakan murni Go standard library (`bufio`, `regexp`, `strings`, `time`).
- **Markdown Checklist Parser**: Mendukung checklist format `- [ ]` / `- [x]`, bullet list, serta numbered list bergradasi.
- **Evidence Gap Analyzer**: Membandingkan korpus bukti laporan (ringkasan, catatan, artefak, output tes) terhadap target kriteria dengan toleransi kecocokan kata kunci (*keyword density matching*).
- **Audit & Metadata Override**: Menerima verifikasi eksplisit lewat metadata untuk pengujian custom.
- **Rekomendasi Aksi Otomatis**: Menghasilkan aksi terarah seperti `mark_done` atau `auto_requeue` sesuai kelulusan persentase DoD.

## Cara Penggunaan

```go
package main

import (
    "fmt"
    "github.com/buchin/dodeval"
)

func main() {
    checklist := `
### Definition of Done
- [ ] Targeted tests pass 100% dalam waktu kurang dari 10 detik
- [ ] Kompilasi build Go 0 error
- [ ] Dokumentasi dan Landing Page diperbarui
`
    evidence := dodeval.Evidence{
        Summary:     "Kompilasi build Go 0 error dan dokumentasi diperbarui.",
        TestSummary: "Targeted tests pass 100% dalam waktu 0.05 detik.",
        Assertions:  []string{"Landing page updated with new capability"},
    }

    evaluator := dodeval.NewEvaluator()
    result := evaluator.EvaluateString(checklist, evidence)

    fmt.Printf("Passed: %v (%.1f%%)\n", result.Passed, result.Percentage)
    fmt.Printf("Action: %s\n", result.Action)
}
```
