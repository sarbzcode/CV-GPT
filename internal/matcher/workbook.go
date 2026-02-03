package matcher

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/xuri/excelize/v2"
)

func ReadWorkbookInputs(workbook string) (Input, error) {
    f, err := excelize.OpenFile(workbook)
    if err != nil {
        return Input{}, err
    }
    defer f.Close()

    jdPath, _ := f.GetCellValue("Inputs", "B1")
    resumesPath, _ := f.GetCellValue("Inputs", "B2")
    topNStr, _ := f.GetCellValue("Inputs", "B3")
    outPath, _ := f.GetCellValue("Inputs", "B4")

    topN := 0
    if topNStr != "" {
        _, _ = fmt.Sscanf(topNStr, "%d", &topN)
    }

    if strings.TrimSpace(outPath) == "" {
        outPath = filepath.Join(filepath.Dir(workbook), "outputs", "results.csv")
    }

    return Input{
        JDPath:     strings.TrimSpace(jdPath),
        ResumesDir: strings.TrimSpace(resumesPath),
        TopN:       topN,
        OutPath:    strings.TrimSpace(outPath),
    }, nil
}
