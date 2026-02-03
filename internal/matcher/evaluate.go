package matcher

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMissingResume = errors.New("resume file not found")
	ErrReadResume    = errors.New("failed to read resume")
)

// EvaluateCandidate generates a GPT-based evaluation for a single resume.
func EvaluateCandidate(jdPath, resumePath string) (ResumeAnalysis, error) {
	LoadDotEnv()
	if strings.TrimSpace(jdPath) == "" || !fileExists(jdPath) {
		return ResumeAnalysis{}, ErrMissingJD
	}
	if strings.TrimSpace(resumePath) == "" || !fileExists(resumePath) {
		return ResumeAnalysis{}, ErrMissingResume
	}

	client, err := newOpenAIClientFromEnv()
	if err != nil {
		return ResumeAnalysis{}, err
	}

	jdRaw, err := extractText(jdPath)
	if err != nil {
		return ResumeAnalysis{}, fmt.Errorf("%w: %v", ErrReadJD, err)
	}

	resumeRaw, err := extractText(resumePath)
	if err != nil {
		return ResumeAnalysis{}, fmt.Errorf("%w: %v", ErrReadResume, err)
	}

	ctx := context.Background()
	jdInfo, err := extractJDInfo(ctx, client, redactPII(jdRaw))
	if err != nil {
		return ResumeAnalysis{}, err
	}

	analysis, err := explainResume(ctx, client, jdInfo, redactPII(resumeRaw))
	if err != nil {
		return ResumeAnalysis{}, err
	}

	return analysis, nil
}
