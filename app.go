package main

import (
	"context"
	"os/exec"
	stdruntime "runtime"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"resume-gpt/internal/matcher"
)

type App struct {
    ctx context.Context
}

func NewApp() *App {
    return &App{}
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
}

func (a *App) SelectJDFile() (string, error) {
    return wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
        Title: "Select Job Description",
        Filters: []wailsruntime.FileFilter{
            {DisplayName: "Documents", Pattern: "*.txt;*.text;*.md;*.pdf;*.docx;*.rtf"},
        },
    })
}

func (a *App) SelectResumesFolder() (string, error) {
    return wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
        Title: "Select Resumes Folder",
    })
}

func (a *App) SelectOutputFile() (string, error) {
	return wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "Save Results CSV",
		DefaultFilename: "results.csv",
		Filters:         []wailsruntime.FileFilter{{DisplayName: "CSV", Pattern: "*.csv"}},
	})
}

func (a *App) RunMatch(jdPath, resumesDir string, topN int, outPath string) (matcher.Output, error) {
    if topN < 0 {
        topN = 0
    }
    input := matcher.Input{
        JDPath:     jdPath,
        ResumesDir: resumesDir,
        TopN:       topN,
        OutPath:    outPath,
    }
    return matcher.RunHeuristic(input)
}

func (a *App) EvaluateCandidate(jdPath, resumePath string) (matcher.ResumeAnalysis, error) {
	return matcher.EvaluateCandidate(jdPath, resumePath)
}

// OpenResumeFile opens the resume in the default OS app.
func (a *App) OpenResumeFile(path string) error {
	return openFile(path)
}

func openFile(path string) error {
	if path == "" {
		return nil
	}
	switch stdruntime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start", "", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}
