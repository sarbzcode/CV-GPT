package main

import (
    "errors"
    "flag"
    "fmt"
    "os"

    "resume-gpt/internal/matcher"
)

func main() {
    matcher.LoadDotEnv()
    workbook := flag.String("workbook", "", "Path to Excel workbook")
    jd := flag.String("jd", "", "Path to job description file")
    resumes := flag.String("resumes", "", "Path to resumes folder")
    topN := flag.Int("topn", 0, "Top N results")
    out := flag.String("out", "", "Output CSV path")
    flag.Parse()

    var input matcher.Input

    if *workbook != "" {
        if _, err := os.Stat(*workbook); err != nil {
            fmt.Fprintln(os.Stderr, "Workbook not found")
            os.Exit(1)
        }
        wbInput, err := matcher.ReadWorkbookInputs(*workbook)
        if err != nil {
            fmt.Fprintln(os.Stderr, "Failed to read workbook:", err)
            os.Exit(1)
        }
        input = wbInput
    } else {
        input = matcher.Input{
            JDPath:     *jd,
            ResumesDir: *resumes,
            TopN:       *topN,
            OutPath:    *out,
        }
        if input.JDPath == "" || input.ResumesDir == "" {
            fmt.Fprintln(os.Stderr, "Provide --workbook or both --jd and --resumes")
            os.Exit(1)
        }
    }

    _, err := matcher.Run(input)
    if err != nil {
        switch {
        case errors.Is(err, matcher.ErrMissingJD):
            fmt.Fprintln(os.Stderr, "Job description file not found")
            os.Exit(2)
        case errors.Is(err, matcher.ErrReadJD):
            fmt.Fprintln(os.Stderr, "Failed to read JD:", err)
            os.Exit(2)
        case errors.Is(err, matcher.ErrMissingResumes):
            fmt.Fprintln(os.Stderr, "Resumes folder not found")
            os.Exit(3)
        case errors.Is(err, matcher.ErrListResumes):
            fmt.Fprintln(os.Stderr, "Failed to list resumes:", err)
            os.Exit(3)
        case errors.Is(err, matcher.ErrNoResumes):
            fmt.Fprintln(os.Stderr, "No resumes found")
            os.Exit(4)
		case errors.Is(err, matcher.ErrWriteResults):
			fmt.Fprintln(os.Stderr, "Failed to write results:", err)
			os.Exit(5)
		case errors.Is(err, matcher.ErrMissingOpenAIKey):
			fmt.Fprintln(os.Stderr, "OpenAI API key not configured (set OPENAI_API_KEY)")
			os.Exit(6)
		default:
			fmt.Fprintln(os.Stderr, "Matcher failed:", err)
			os.Exit(5)
		}
	}

    fmt.Fprintln(os.Stdout, "Done")
}
