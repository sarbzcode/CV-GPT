# ResumeGPT (CV-GPT)

ResumeGPT ranks resumes against a job description using a Go matcher, with three ways to run it:
- CLI (`cmd/resume_matcher`)
- Excel macro workbook (`ResumeMatcher.xlsm` + `vba/ResumeMatcher.bas`)
- Desktop app (Wails, `main.go` + `app.go`)

## Features
- Supports `.txt`, `.md`, `.pdf`, `.docx`, `.rtf` inputs
- Scores and ranks candidates with strengths/weaknesses
- Writes `results.csv` and `run_log.txt`
- Optional OpenAI mode for semantic ranking and richer explanations
- Optional per-candidate AI evaluation in the desktop UI

## How the system works
1. Input stage: JD file + resumes folder + optional Top N/output path are provided from CLI, Excel, or desktop UI.
2. Parsing stage: files are read and converted to text (`internal/matcher/matcher.go`), then PII/demographic terms are redacted before scoring.
3. Ranking stage:
   - **Heuristic mode**: TF-IDF cosine similarity + skill matching (must/nice/general) with weighted scoring.
   - **OpenAI mode** (if `OPENAI_API_KEY` is set): extracts structured JD requirements, embeds JD/resumes, computes similarity + skill coverage, then generates top-N explanations.
4. Output stage: results are sorted, optionally truncated to Top N, and written to CSV.
5. Display stage: Excel imports CSV into `Results` sheet; desktop UI renders table and can generate candidate-specific evaluations.

## Project structure
- `cmd/resume_matcher/main.go` - CLI entry point
- `internal/matcher/` - core parsing, scoring, OpenAI integration, workbook input reader
- `app.go`, `main.go` - Wails desktop bindings and app bootstrap
- `frontend/dist/` - desktop UI assets currently embedded into the app
- `vba/ResumeMatcher.bas` - Excel automation macros
- `samples/` - sample JD/resume files

## Prerequisites
- Go `1.24+`
- Windows + Excel (only if you want the Excel UI)
- Wails CLI (only if you want desktop dev/build):  
  `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Environment setup
1. Create your local env file:
   ```powershell
   Copy-Item .env.example .env
   ```
2. Edit `.env` and set at least:
   ```env
   OPENAI_API_KEY=your_openai_api_key_here
   ```
3. OpenAI is optional unless you set:
   ```env
   RESUMEGPT_REQUIRE_OPENAI=1
   ```

## Run options

### 1) CLI
Build:
```powershell
go build -o bin\resume_matcher.exe .\cmd\resume_matcher
```

Run:
```powershell
bin\resume_matcher.exe --jd path\to\jd.pdf --resumes path\to\resumes --topn 25 --out outputs\results.csv
```

### 2) Desktop app (Wails)
Dev:
```powershell
wails dev
```

Build:
```powershell
wails build
```

### 3) Excel workbook
1. Open `ResumeMatcher.xlsm`.
2. Ensure macros are enabled.
3. If needed, import `vba/ResumeMatcher.bas` in VBA editor.
4. Set:
   - `Inputs!B1` = JD file
   - `Inputs!B2` = resumes folder
   - `Inputs!B3` = Top N (optional)
   - `Inputs!B7` = matcher exe path (optional; defaults to `bin\resume_matcher.exe`)
   - `Inputs!B8` = project root (optional, if workbook is outside project)
5. Click **Run Matcher**.

