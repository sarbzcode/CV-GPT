package matcher

import (
    "bufio"
    "context"
    "encoding/csv"
    "encoding/json"
    "errors"
    "fmt"
    "math"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strings"
    "time"

    "github.com/ledongthuc/pdf"
    "github.com/nguyenthenguyen/docx"
)

type Result struct {
    Rank        int     `json:"rank"`
    Candidate   string  `json:"candidate"`
    Score       float64 `json:"score"`
    Strengths   string  `json:"strengths"`
    Weaknesses  string  `json:"weaknesses"`
    Explanation string  `json:"explanation"`
    File        string  `json:"file"`
    Extracted   *ResumeExtract `json:"extracted,omitempty"`
}

type Input struct {
    JDPath     string
    ResumesDir string
    TopN       int
    OutPath    string
}

type Output struct {
    Results []Result `json:"results"`
    OutPath string   `json:"outPath"`
    Total   int      `json:"total"`
    JDInfo  *JDExtract `json:"jdInfo,omitempty"`
}

type JDExtract struct {
    RoleTitle          string   `json:"role_title"`
    SkillsMust         []string `json:"skills_must"`
    SkillsNice         []string `json:"skills_nice"`
    SkillsOther        []string `json:"skills_other"`
    YearsExperienceMin float64  `json:"years_experience_min"`
    Education          []string `json:"education"`
    Certifications     []string `json:"certifications"`
    Titles             []string `json:"titles"`
    Responsibilities   []string `json:"responsibilities"`
}

type ResumeExtract struct {
    Skills          []string `json:"skills"`
    YearsExperience float64  `json:"years_experience"`
    Education       []string `json:"education"`
    Certifications  []string `json:"certifications"`
    Titles          []string `json:"titles"`
}

type ResumeAnalysis struct {
    Strengths       []string `json:"strengths"`
    Weaknesses      []string `json:"weaknesses"`
    Summary         string   `json:"summary"`
    Skills          []string `json:"skills"`
    YearsExperience float64  `json:"years_experience"`
    Education       []string `json:"education"`
    Certifications  []string `json:"certifications"`
    Titles          []string `json:"titles"`
}

type resumeDoc struct {
    Path     string
    Name     string
    Raw      string
    Redacted string
    Norm     string
}

var stopwords = map[string]bool{
    "a": true, "an": true, "the": true, "and": true, "or": true, "but": true, "if": true,
    "then": true, "else": true, "for": true, "to": true, "of": true, "in": true, "on": true,
    "at": true, "by": true, "with": true, "from": true, "as": true, "is": true, "are": true,
    "was": true, "were": true, "be": true, "been": true, "this": true, "that": true,
    "these": true, "those": true, "it": true, "its": true, "we": true, "our": true,
    "you": true, "your": true, "they": true, "their": true, "i": true, "me": true, "my": true,
    "he": true, "she": true, "him": true, "her": true, "them": true, "us": true, "can": true,
    "could": true, "should": true, "would": true, "will": true, "may": true, "might": true,
    "not": true, "no": true, "yes": true, "do": true, "does": true, "did": true,
}

var redactTerms = []string{
    "male", "female", "man", "woman", "men", "women", "boy", "girl", "mr", "mrs", "ms",
    "he", "she", "him", "her", "his", "hers", "mother", "father", "husband", "wife",
    "married", "single", "divorced", "age", "aged", "years old", "birthday",
    "religion", "christian", "muslim", "hindu", "jewish", "buddhist", "sikh",
    "white", "black", "asian", "latino", "hispanic", "native", "indigenous",
    "citizenship", "nationality", "veteran", "disability", "disabled",
}

var skillLexicon = []string{
    "python", "java", "c", "c++", "c#", "go", "golang", "rust", "scala", "kotlin",
    "swift", "objective-c", "javascript", "typescript", "ruby", "php", "perl",
    "matlab", "r", "sas", "stata", "julia", "sql", "pl/sql", "t-sql", "nosql",
    "html", "css", "sass", "less", "json", "xml", "yaml", "graphql", "rest", "grpc",
    "api", "microservices", "soa", "oop", "design patterns", "clean architecture",
    "react", "react.js", "angular", "vue", "svelte", "next.js", "nuxt", "node.js",
    "nodejs", "express", "nestjs", "django", "flask", "fastapi", "spring", "spring boot",
    "asp.net", ".net", "entity framework", "laravel", "rails", "ruby on rails",
    "gin", "echo", "fiber", "wails", "electron", "qt",
    "android", "ios", "react native", "flutter", "xamarin", "cordova",
    "aws", "amazon web services", "azure", "gcp", "google cloud", "oracle cloud",
    "docker", "kubernetes", "helm", "terraform", "ansible", "chef", "puppet",
    "jenkins", "github actions", "gitlab ci", "circleci", "ci/cd", "devops",
    "linux", "windows", "macos", "bash", "powershell", "shell scripting",
    "git", "svn", "mercurial",
    "postgresql", "mysql", "mariadb", "sql server", "oracle", "sqlite", "mongodb",
    "cassandra", "redis", "dynamodb", "elasticsearch", "opensearch", "neo4j",
    "snowflake", "bigquery", "redshift", "databricks",
    "kafka", "rabbitmq", "activemq", "nats", "sqs", "pubsub",
    "spark", "hadoop", "hive", "pig", "airflow", "dbt", "etl", "elt", "data pipeline",
    "data warehouse", "data lake", "data modeling", "data governance",
    "machine learning", "deep learning", "nlp", "computer vision", "llm",
    "data analysis", "data analytics", "data science", "statistics",
    "feature engineering", "modeling", "forecasting", "recommendation systems",
    "pandas", "numpy", "scikit-learn", "tensorflow", "pytorch", "keras", "xgboost",
    "lightgbm", "catboost", "mlops", "model deployment", "onnx",
    "excel", "power bi", "tableau", "looker", "qlik", "superset", "mode",
    "salesforce", "sap", "oracle erp", "netsuite", "workday",
    "serviceNow", "jira", "confluence", "slack", "microsoft teams",
    "testing", "unit testing", "integration testing", "e2e testing", "tdd", "bdd",
    "jest", "mocha", "cypress", "playwright", "selenium", "pytest", "junit",
    "security", "oauth", "openid connect", "saml", "jwt", "encryption",
    "identity", "iam", "zero trust", "vulnerability management",
    "networking", "tcp/ip", "dns", "http", "https", "ssl", "tls", "load balancing",
    "observability", "monitoring", "logging", "tracing", "prometheus", "grafana",
    "datadog", "new relic", "splunk",
    "product management", "project management", "agile", "scrum", "kanban",
    "leadership", "stakeholder management", "communication", "requirements",
    "documentation", "technical writing",
    "ui/ux", "figma", "sketch", "adobe xd", "user research", "wireframing",
    "seo", "marketing", "growth", "analytics", "a/b testing",
    "accounting", "finance", "budgeting", "forecasting", "procurement",
    "hr", "recruiting", "talent acquisition", "payroll", "benefits",
    "customer support", "sales", "business development", "crm",
    "compliance", "risk management", "gdpr", "hipaa", "sox", "pci",
    "warehouse", "logistics", "supply chain", "operations",
}

var emailRe = regexp.MustCompile(`[\w\.-]+@[\w\.-]+`)
var phoneRe = regexp.MustCompile(`\+?\d[\d\s\-]{7,}`)
var nonWordRe = regexp.MustCompile(`[^a-z0-9\s\+#]`)

var (
    ErrMissingJD      = errors.New("job description file not found")
    ErrMissingResumes = errors.New("resumes folder not found")
    ErrNoResumes      = errors.New("no resumes found")
    ErrReadJD         = errors.New("failed to read JD")
    ErrListResumes    = errors.New("failed to list resumes")
    ErrWriteResults   = errors.New("failed to write results")
    ErrMissingOpenAIKey = errors.New("openai api key not configured")
)

func Run(input Input) (Output, error) {
    return runInternal(input, false)
}

// RunHeuristic always uses the legacy heuristic ranking (no OpenAI).
func RunHeuristic(input Input) (Output, error) {
    return runInternal(input, true)
}

func runInternal(input Input, forceHeuristic bool) (Output, error) {
    LoadDotEnv()
    if strings.TrimSpace(input.JDPath) == "" || !fileExists(input.JDPath) {
        return Output{}, ErrMissingJD
    }
    if strings.TrimSpace(input.ResumesDir) == "" || !dirExists(input.ResumesDir) {
        return Output{}, ErrMissingResumes
    }

    jdRaw, err := extractText(input.JDPath)
    if err != nil {
        return Output{}, fmt.Errorf("%w: %v", ErrReadJD, err)
    }

    resumeFiles, err := listResumeFiles(input.ResumesDir)
    if err != nil {
        return Output{}, fmt.Errorf("%w: %v", ErrListResumes, err)
    }
    if len(resumeFiles) == 0 {
        return Output{}, ErrNoResumes
    }
    totalResumes := len(resumeFiles)

    resumeDocs := make([]resumeDoc, 0, len(resumeFiles))
    for _, path := range resumeFiles {
        raw, err := extractText(path)
        if err != nil {
            continue
        }
        resumeDocs = append(resumeDocs, resumeDoc{
            Path:     path,
            Name:     strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
            Raw:      raw,
            Redacted: redactPII(raw),
            Norm:     normalizeText(raw),
        })
    }
    if len(resumeDocs) == 0 {
        return Output{}, ErrNoResumes
    }

    if forceHeuristic {
        return runHeuristic(input, jdRaw, resumeDocs, totalResumes)
    }

    aiRequired := envBool("RESUMEGPT_REQUIRE_OPENAI", false)
    aiClient, aiErr := newOpenAIClientFromEnv()
    if aiErr == nil {
        return runOpenAI(input, jdRaw, resumeDocs, totalResumes, aiClient)
    }
    if aiRequired {
        if errors.Is(aiErr, ErrMissingOpenAIKey) {
            return Output{}, ErrMissingOpenAIKey
        }
        return Output{}, aiErr
    }
    return runHeuristic(input, jdRaw, resumeDocs, totalResumes)
}

func runHeuristic(input Input, jdRaw string, resumeDocs []resumeDoc, totalResumes int) (Output, error) {
    jdNorm := normalizeText(jdRaw)
    jdTerms := topTerms(jdNorm, 25)
    jdSkills := extractSkills(jdNorm, jdTerms)
    mustSkills, niceSkills := findMustNiceSkills(jdRaw)

    docs := []string{jdNorm}
    resumeTexts := make([]string, 0, len(resumeDocs))
    resumeNames := make([]string, 0, len(resumeDocs))
    resumeFiles := make([]string, 0, len(resumeDocs))
    for _, doc := range resumeDocs {
        docs = append(docs, doc.Norm)
        resumeTexts = append(resumeTexts, doc.Norm)
        resumeNames = append(resumeNames, doc.Name)
        resumeFiles = append(resumeFiles, doc.Path)
    }

    vectors := buildTfidfVectors(docs)
    jdVec := vectors[0]

    wCos, wMust, wNice, wSkill := scoreWeights(len(mustSkills))

    results := make([]Result, 0, len(resumeTexts))
    for i, text := range resumeTexts {
        resSkills := extractSkills(text, jdTerms)
        resSkillSet := map[string]bool{}
        for _, s := range resSkills {
            resSkillSet[s] = true
        }

        mustMatch := 0
        for _, s := range mustSkills {
            if resSkillSet[s] {
                mustMatch++
            }
        }
        niceMatch := 0
        for _, s := range niceSkills {
            if resSkillSet[s] {
                niceMatch++
            }
        }
        skillMatch := 0
        for _, s := range jdSkills {
            if resSkillSet[s] {
                skillMatch++
            }
        }

        mustRatio := ratio(mustMatch, len(mustSkills))
        niceRatio := ratio(niceMatch, len(niceSkills))
        skillRatio := ratio(skillMatch, len(jdSkills))

        sim := cosineSim(jdVec, vectors[i+1])

        score := (wCos * sim) + (wMust * mustRatio) + (wNice * niceRatio) + (wSkill * skillRatio)
        scorePct := round(score * 100)

        strengths := buildStrengths(resSkillSet, mustSkills, niceSkills, jdSkills)
        weaknesses := buildWeaknesses(resSkillSet, mustSkills, niceSkills, jdSkills)

        explanation := fmt.Sprintf(
            "Similarity=%.2f; MustMatch=%.2f; NiceMatch=%.2f; SkillMatch=%.2f",
            sim, mustRatio, niceRatio, skillRatio,
        )

        results = append(results, Result{
            Candidate:   resumeNames[i],
            Score:       scorePct,
            Strengths:   joinOrNone(strengths),
            Weaknesses:  joinOrNone(weaknesses),
            Explanation: explanation,
            File:        resumeFiles[i],
        })
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })
    for i := range results {
        results[i].Rank = i + 1
    }

    if input.TopN > 0 && len(results) > input.TopN {
        results = results[:input.TopN]
    }

    outPath := strings.TrimSpace(input.OutPath)
    if outPath == "" {
        outPath = filepath.Join("outputs", "results.csv")
    }

    if err := writeResultsCSV(outPath, results); err != nil {
        return Output{}, fmt.Errorf("%w: %v", ErrWriteResults, err)
    }
    appendLog(outPath, totalResumes)

    return Output{Results: results, OutPath: outPath, Total: totalResumes}, nil
}

func runOpenAI(input Input, jdRaw string, resumeDocs []resumeDoc, totalResumes int, client *openAIClient) (Output, error) {
    ctx := context.Background()
    jdRedacted := redactPII(jdRaw)
    jdNorm := normalizeText(jdRaw)

    jdInfo, err := extractJDInfo(ctx, client, jdRedacted)
    if err != nil {
        return Output{}, err
    }

    jdTerms := topTerms(jdNorm, 25)
    fallbackSkills := extractSkills(jdNorm, jdTerms)

    mustSkills := cleanSkillList(jdInfo.SkillsMust)
    niceSkills := cleanSkillList(jdInfo.SkillsNice)
    otherSkills := cleanSkillList(jdInfo.SkillsOther)

    if len(mustSkills) == 0 && len(niceSkills) == 0 && len(otherSkills) == 0 {
        mustSkills, niceSkills = findMustNiceSkills(jdRaw)
        otherSkills = fallbackSkills
    }

    allSkills := mergeUnique(mustSkills, niceSkills, otherSkills, fallbackSkills)

    docTexts := make([]string, 0, len(resumeDocs)+1)
    docTexts = append(docTexts, jdRedacted)
    for _, doc := range resumeDocs {
        docTexts = append(docTexts, doc.Redacted)
    }
    embeddings, err := embedDocuments(ctx, client, docTexts)
    if err != nil {
        return Output{}, err
    }
    jdVec := embeddings[0]

    wCos, wMust, wNice, wSkill := scoreWeights(len(mustSkills))

    results := make([]Result, 0, len(resumeDocs))
    resumeByPath := make(map[string]resumeDoc, len(resumeDocs))
    for i, doc := range resumeDocs {
        resumeByPath[doc.Path] = doc
        vec := embeddings[i+1]

        resSkillSet := skillsInText(doc.Norm, allSkills)
        mustMatch := countMatches(resSkillSet, mustSkills)
        niceMatch := countMatches(resSkillSet, niceSkills)
        skillMatch := countMatches(resSkillSet, allSkills)

        mustRatio := ratio(mustMatch, len(mustSkills))
        niceRatio := ratio(niceMatch, len(niceSkills))
        skillRatio := ratio(skillMatch, len(allSkills))

        sim := cosineSimVec(jdVec, vec)

        score := (wCos * sim) + (wMust * mustRatio) + (wNice * niceRatio) + (wSkill * skillRatio)
        scorePct := round(score * 100)

        strengths := buildStrengths(resSkillSet, mustSkills, niceSkills, allSkills)
        weaknesses := buildWeaknesses(resSkillSet, mustSkills, niceSkills, allSkills)

        explanation := fmt.Sprintf(
            "Similarity=%.2f; MustMatch=%.2f; NiceMatch=%.2f; SkillMatch=%.2f",
            sim, mustRatio, niceRatio, skillRatio,
        )

        results = append(results, Result{
            Candidate:   doc.Name,
            Score:       scorePct,
            Strengths:   joinOrNone(strengths),
            Weaknesses:  joinOrNone(weaknesses),
            Explanation: explanation,
            File:        doc.Path,
        })
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })
    for i := range results {
        results[i].Rank = i + 1
    }

    if input.TopN > 0 && len(results) > input.TopN {
        results = results[:input.TopN]
    }

    explainN := client.explainTopN
    if input.TopN > 0 && input.TopN < explainN {
        explainN = input.TopN
    }
    if explainN > len(results) {
        explainN = len(results)
    }

    for i := 0; i < explainN; i++ {
        doc, ok := resumeByPath[results[i].File]
        if !ok {
            continue
        }
        analysis, err := explainResume(ctx, client, jdInfo, doc.Redacted)
        if err != nil {
            continue
        }
        results[i].Strengths = joinOrNone(analysis.Strengths)
        results[i].Weaknesses = joinOrNone(analysis.Weaknesses)
        if strings.TrimSpace(analysis.Summary) != "" {
            results[i].Explanation = analysis.Summary
        }
        results[i].Extracted = &ResumeExtract{
            Skills:          cleanSkillList(analysis.Skills),
            YearsExperience: analysis.YearsExperience,
            Education:       cleanList(analysis.Education),
            Certifications:  cleanList(analysis.Certifications),
            Titles:          cleanList(analysis.Titles),
        }
    }

    outPath := strings.TrimSpace(input.OutPath)
    if outPath == "" {
        outPath = filepath.Join("outputs", "results.csv")
    }

    if err := writeResultsCSV(outPath, results); err != nil {
        return Output{}, fmt.Errorf("%w: %v", ErrWriteResults, err)
    }
    appendLog(outPath, totalResumes)

    return Output{Results: results, OutPath: outPath, Total: totalResumes, JDInfo: &jdInfo}, nil
}

func extractJDInfo(ctx context.Context, client *openAIClient, jdText string) (JDExtract, error) {
    system := strings.Join([]string{
        "You extract only job-related requirements.",
        "Ignore demographics or personal details.",
        "If a field is missing, return empty arrays or 0.",
        "Return only JSON that matches the schema.",
    }, " ")
    schema := jdExtractSchema()
    var out JDExtract
    if err := client.chatCompletionJSON(ctx, "jd_extract", schema, system, jdText, &out); err != nil {
        return JDExtract{}, err
    }
    out.RoleTitle = strings.TrimSpace(out.RoleTitle)
    out.SkillsMust = cleanSkillList(out.SkillsMust)
    out.SkillsNice = cleanSkillList(out.SkillsNice)
    out.SkillsOther = cleanSkillList(out.SkillsOther)
    out.Education = cleanList(out.Education)
    out.Certifications = cleanList(out.Certifications)
    out.Titles = cleanList(out.Titles)
    out.Responsibilities = cleanList(out.Responsibilities)
    if out.YearsExperienceMin < 0 {
        out.YearsExperienceMin = 0
    }
    return out, nil
}

func explainResume(ctx context.Context, client *openAIClient, jd JDExtract, resumeText string) (ResumeAnalysis, error) {
    system := strings.Join([]string{
        "You evaluate a resume against job requirements.",
        "Focus only on job-relevant skills and experience.",
        "Ignore names, demographics, and personal details.",
        "Return only JSON that matches the schema.",
    }, " ")

    jdJSON, _ := json.Marshal(jd)
    trimmed := truncateText(resumeText, client.explainMaxChars)
    user := fmt.Sprintf("Job requirements JSON:\n%s\n\nResume:\n%s", string(jdJSON), trimmed)

    schema := resumeAnalysisSchema()
    var out ResumeAnalysis
    if err := client.chatCompletionJSON(ctx, "resume_analysis", schema, system, user, &out); err != nil {
        return ResumeAnalysis{}, err
    }
    out.Strengths = cleanList(out.Strengths)
    out.Weaknesses = cleanList(out.Weaknesses)
    out.Skills = cleanSkillList(out.Skills)
    out.Education = cleanList(out.Education)
    out.Certifications = cleanList(out.Certifications)
    out.Titles = cleanList(out.Titles)
    if out.YearsExperience < 0 {
        out.YearsExperience = 0
    }
    return out, nil
}

func embedDocuments(ctx context.Context, client *openAIClient, docs []string) ([][]float64, error) {
    chunks := make([]string, 0)
    docChunkIdxs := make([][]int, len(docs))
    for i, doc := range docs {
        parts := chunkByWords(doc, client.embedChunkWords)
        for _, part := range parts {
            idx := len(chunks)
            chunks = append(chunks, part)
            docChunkIdxs[i] = append(docChunkIdxs[i], idx)
        }
    }
    if len(chunks) == 0 {
        return make([][]float64, len(docs)), nil
    }
    embeds, err := client.embedTexts(ctx, chunks)
    if err != nil {
        return nil, err
    }
    out := make([][]float64, len(docs))
    for i, idxs := range docChunkIdxs {
        vecs := make([][]float64, 0, len(idxs))
        for _, idx := range idxs {
            if idx >= 0 && idx < len(embeds) && len(embeds[idx]) > 0 {
                vecs = append(vecs, embeds[idx])
            }
        }
        out[i] = averageEmbeddings(vecs)
    }
    return out, nil
}

func chunkByWords(text string, wordsPerChunk int) []string {
    tokens := strings.Fields(strings.TrimSpace(text))
    if len(tokens) == 0 {
        return []string{}
    }
    if len(tokens) <= wordsPerChunk {
        return []string{strings.Join(tokens, " ")}
    }
    chunks := make([]string, 0)
    for i := 0; i < len(tokens); i += wordsPerChunk {
        end := i + wordsPerChunk
        if end > len(tokens) {
            end = len(tokens)
        }
        chunks = append(chunks, strings.Join(tokens[i:end], " "))
    }
    return chunks
}

func averageEmbeddings(vecs [][]float64) []float64 {
    if len(vecs) == 0 {
        return nil
    }
    dim := len(vecs[0])
    if dim == 0 {
        return nil
    }
    sum := make([]float64, dim)
    count := 0.0
    for _, v := range vecs {
        if len(v) != dim {
            continue
        }
        for i := 0; i < dim; i++ {
            sum[i] += v[i]
        }
        count++
    }
    if count == 0 {
        return nil
    }
    for i := range sum {
        sum[i] /= count
    }
    return sum
}

func cosineSimVec(a, b []float64) float64 {
    if len(a) == 0 || len(b) == 0 {
        return 0
    }
    dim := len(a)
    if len(b) < dim {
        dim = len(b)
    }
    dot := 0.0
    normA := 0.0
    normB := 0.0
    for i := 0; i < dim; i++ {
        dot += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    if normA == 0 || normB == 0 {
        return 0
    }
    return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func skillsInText(text string, skills []string) map[string]bool {
    set := map[string]bool{}
    for _, s := range skills {
        if s == "" {
            continue
        }
        if strings.Contains(text, s) {
            set[s] = true
        }
    }
    return set
}

func countMatches(set map[string]bool, list []string) int {
    count := 0
    for _, s := range list {
        if set[s] {
            count++
        }
    }
    return count
}

func cleanSkillList(items []string) []string {
    cleaned := make([]string, 0, len(items))
    seen := map[string]bool{}
    for _, item := range items {
        t := strings.ToLower(strings.TrimSpace(item))
        if t == "" || seen[t] {
            continue
        }
        seen[t] = true
        cleaned = append(cleaned, t)
    }
    sort.Strings(cleaned)
    return cleaned
}

func cleanList(items []string) []string {
    cleaned := make([]string, 0, len(items))
    seen := map[string]bool{}
    for _, item := range items {
        t := strings.TrimSpace(item)
        if t == "" || seen[t] {
            continue
        }
        seen[t] = true
        cleaned = append(cleaned, t)
    }
    sort.Strings(cleaned)
    return cleaned
}

func mergeUnique(lists ...[]string) []string {
    out := make([]string, 0)
    seen := map[string]bool{}
    for _, list := range lists {
        for _, item := range list {
            t := strings.TrimSpace(item)
            if t == "" || seen[t] {
                continue
            }
            seen[t] = true
            out = append(out, t)
        }
    }
    sort.Strings(out)
    return out
}

func truncateText(text string, maxChars int) string {
    if maxChars <= 0 {
        return text
    }
    runes := []rune(text)
    if len(runes) <= maxChars {
        return text
    }
    return string(runes[:maxChars])
}

func jdExtractSchema() map[string]any {
    return map[string]any{
        "type":                 "object",
        "additionalProperties": false,
        "properties": map[string]any{
            "role_title": map[string]any{"type": "string"},
            "skills_must": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
            "skills_nice": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
            "skills_other": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
            "years_experience_min": map[string]any{
                "type":    "number",
                "minimum": 0,
            },
            "education": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
            "certifications": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
            "titles": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
            "responsibilities": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
        },
        "required": []string{
            "role_title",
            "skills_must",
            "skills_nice",
            "skills_other",
            "years_experience_min",
            "education",
            "certifications",
            "titles",
            "responsibilities",
        },
    }
}

func resumeAnalysisSchema() map[string]any {
    return map[string]any{
        "type":                 "object",
        "additionalProperties": false,
        "properties": map[string]any{
            "strengths": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
            "weaknesses": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
            "summary": map[string]any{"type": "string"},
            "skills": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
            "years_experience": map[string]any{
                "type":    "number",
                "minimum": 0,
            },
            "education": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
            "certifications": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
            "titles": map[string]any{
                "type":  "array",
                "items": map[string]any{"type": "string"},
            },
        },
        "required": []string{
            "strengths",
            "weaknesses",
            "summary",
            "skills",
            "years_experience",
            "education",
            "certifications",
            "titles",
        },
    }
}

func readTxt(path string) (string, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    return string(b), nil
}

func readDocx(path string) (string, error) {
    doc, err := docx.ReadDocxFile(path)
    if err != nil {
        return "", err
    }
    defer doc.Close()
    return doc.Editable().GetContent(), nil
}

func readPdf(path string) (string, error) {
    f, r, err := pdf.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()

    var sb strings.Builder
    totalPage := r.NumPage()
    for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
        p := r.Page(pageIndex)
        if p.V.IsNull() {
            continue
        }
        txt, err := p.GetPlainText(nil)
        if err == nil {
            sb.WriteString(txt)
            sb.WriteString("\n")
        }
    }
    return sb.String(), nil
}

func readRtf(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }

    var out strings.Builder
    for i := 0; i < len(data); i++ {
        b := data[i]
        switch b {
        case '{', '}':
            continue
        case '\\':
            if i+1 >= len(data) {
                continue
            }
            next := data[i+1]
            if next == '\'' && i+3 < len(data) {
                hi := data[i+2]
                lo := data[i+3]
                val, ok := rtfHex(hi, lo)
                if ok {
                    out.WriteByte(val)
                }
                i += 3
                continue
            }

            nameStart := i + 1
            j := nameStart
            for j < len(data) && ((data[j] >= 'a' && data[j] <= 'z') || (data[j] >= 'A' && data[j] <= 'Z')) {
                j++
            }
            name := strings.ToLower(string(data[nameStart:j]))

            sign := 1
            if j < len(data) && data[j] == '-' {
                sign = -1
                j++
            }
            numStart := j
            for j < len(data) && data[j] >= '0' && data[j] <= '9' {
                j++
            }
            param := 0
            hasNum := numStart < j
            if hasNum {
                for k := numStart; k < j; k++ {
                    param = param*10 + int(data[k]-'0')
                }
                param *= sign
            }

            if name == "u" && hasNum {
                r := rune(param)
                if r < 0 {
                    r += 65536
                }
                out.WriteRune(r)
                if j < len(data) {
                    j++
                }
            }

            if j < len(data) && data[j] == ' ' {
                i = j
            } else {
                i = j - 1
            }
        default:
            if b >= 32 || b == '\n' || b == '\t' {
                out.WriteByte(b)
            }
        }
    }
    return out.String(), nil
}

func extractText(path string) (string, error) {
    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
    case ".txt", ".text", ".md":
        return readTxt(path)
    case ".docx":
        return readDocx(path)
    case ".pdf":
        return readPdf(path)
    case ".rtf":
        return readRtf(path)
    default:
        return "", fmt.Errorf("unsupported file: %s", ext)
    }
}

func normalizeText(text string) string {
    t := strings.ToLower(redactPII(text))
    t = nonWordRe.ReplaceAllString(t, " ")

    tokens := strings.Fields(t)
    out := make([]string, 0, len(tokens))
    for _, tok := range tokens {
        if !stopwords[tok] {
            out = append(out, tok)
        }
    }
    return strings.Join(out, " ")
}

func redactPII(text string) string {
    t := emailRe.ReplaceAllString(text, " ")
    t = phoneRe.ReplaceAllString(t, " ")
    for _, term := range redactTerms {
        re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(term) + `\b`)
        t = re.ReplaceAllString(t, " ")
    }
    return t
}

func buildNgrams(tokens []string, n int) []string {
    if n <= 1 {
        return tokens
    }
    grams := make([]string, 0)
    for i := 0; i+n <= len(tokens); i++ {
        grams = append(grams, strings.Join(tokens[i:i+n], " "))
    }
    return grams
}

func buildTfidfVectors(docs []string) []map[string]float64 {
    tokenized := make([][]string, 0, len(docs))
    df := map[string]int{}

    for _, doc := range docs {
        tokens := strings.Fields(doc)
        grams := make([]string, 0, len(tokens)*2)
        grams = append(grams, buildNgrams(tokens, 1)...)
        grams = append(grams, buildNgrams(tokens, 2)...)
        tokenized = append(tokenized, grams)

        seen := map[string]bool{}
        for _, g := range grams {
            if !seen[g] {
                df[g]++
                seen[g] = true
            }
        }
    }

    nDocs := float64(len(docs))
    idf := map[string]float64{}
    for term, count := range df {
        idf[term] = math.Log((1+nDocs)/(1+float64(count))) + 1
    }

    vectors := make([]map[string]float64, 0, len(docs))
    for _, grams := range tokenized {
        tf := map[string]float64{}
        for _, g := range grams {
            tf[g]++
        }
        maxTf := 1.0
        for _, v := range tf {
            if v > maxTf {
                maxTf = v
            }
        }
        vec := map[string]float64{}
        for g, v := range tf {
            vec[g] = (v / maxTf) * idf[g]
        }
        vectors = append(vectors, vec)
    }
    return vectors
}

func cosineSim(a, b map[string]float64) float64 {
    if len(a) == 0 || len(b) == 0 {
        return 0
    }
    dot := 0.0
    for k, v := range a {
        dot += v * b[k]
    }
    normA := 0.0
    for _, v := range a {
        normA += v * v
    }
    normB := 0.0
    for _, v := range b {
        normB += v * v
    }
    if normA == 0 || normB == 0 {
        return 0
    }
    return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func topTerms(text string, maxTerms int) []string {
    tokens := strings.Fields(text)
    freq := map[string]int{}
    for _, t := range tokens {
        if stopwords[t] {
            continue
        }
        freq[t]++
    }
    type kv struct {
        k string
        v int
    }
    pairs := make([]kv, 0, len(freq))
    for k, v := range freq {
        pairs = append(pairs, kv{k, v})
    }
    sort.Slice(pairs, func(i, j int) bool {
        if pairs[i].v == pairs[j].v {
            return pairs[i].k < pairs[j].k
        }
        return pairs[i].v > pairs[j].v
    })
    if len(pairs) > maxTerms {
        pairs = pairs[:maxTerms]
    }
    out := make([]string, 0, len(pairs))
    for _, p := range pairs {
        out = append(out, p.k)
    }
    return out
}

func extractSkills(text string, jdTerms []string) []string {
    skills := map[string]bool{}
    for _, s := range skillLexicon {
        if strings.Contains(text, s) {
            skills[s] = true
        }
    }
    for _, t := range jdTerms {
        if len(t) >= 3 && strings.Contains(text, t) {
            skills[t] = true
        }
    }
    out := make([]string, 0, len(skills))
    for s := range skills {
        out = append(out, s)
    }
    sort.Strings(out)
    return out
}

func findMustNiceSkills(jdRaw string) (must []string, nice []string) {
    lines := strings.Split(strings.ToLower(jdRaw), "\n")
    mustSet := map[string]bool{}
    niceSet := map[string]bool{}

    mustKeys := []string{"must", "required", "minimum", "mandatory"}
    niceKeys := []string{"nice to have", "preferred", "plus", "bonus", "optional"}

    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        isMust := false
        for _, k := range mustKeys {
            if strings.Contains(line, k) {
                isMust = true
                break
            }
        }
        isNice := false
        for _, k := range niceKeys {
            if strings.Contains(line, k) {
                isNice = true
                break
            }
        }
        if !isMust && !isNice {
            continue
        }
        for _, s := range skillLexicon {
            if strings.Contains(line, s) {
                if isMust {
                    mustSet[s] = true
                }
                if isNice {
                    niceSet[s] = true
                }
            }
        }
    }

    for s := range mustSet {
        must = append(must, s)
    }
    for s := range niceSet {
        nice = append(nice, s)
    }
    sort.Strings(must)
    sort.Strings(nice)
    return must, nice
}

func scoreWeights(mustCount int) (float64, float64, float64, float64) {
    wCos := 0.45
    wMust := 0.35
    wNice := 0.10
    wSkill := 0.10
    if mustCount == 0 {
        wCos = 0.55
        wNice = 0.15
        wSkill = 0.30
        wMust = 0.0
    }
    return wCos, wMust, wNice, wSkill
}

func listResumeFiles(dir string) ([]string, error) {
    files := []string{}
    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }
        ext := strings.ToLower(filepath.Ext(path))
        if ext == ".txt" || ext == ".text" || ext == ".md" || ext == ".pdf" || ext == ".docx" || ext == ".rtf" {
            files = append(files, path)
        }
        return nil
    })
    return files, err
}

func rtfHex(hi, lo byte) (byte, bool) {
    hv, ok1 := hexVal(hi)
    lv, ok2 := hexVal(lo)
    if !ok1 || !ok2 {
        return 0, false
    }
    return (hv << 4) | lv, true
}

func hexVal(b byte) (byte, bool) {
    switch {
    case b >= '0' && b <= '9':
        return b - '0', true
    case b >= 'a' && b <= 'f':
        return b - 'a' + 10, true
    case b >= 'A' && b <= 'F':
        return b - 'A' + 10, true
    default:
        return 0, false
    }
}

func writeResultsCSV(path string, results []Result) error {
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return err
    }

    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    w := csv.NewWriter(f)
    _ = w.Write([]string{"Rank", "Candidate", "Score", "Strengths", "Weaknesses", "Explanation", "File"})
    for _, r := range results {
        _ = w.Write([]string{
            fmt.Sprintf("%d", r.Rank),
            r.Candidate,
            fmt.Sprintf("%.2f", r.Score),
            r.Strengths,
            r.Weaknesses,
            r.Explanation,
            r.File,
        })
    }
    w.Flush()
    return w.Error()
}

func appendLog(outPath string, total int) {
    logPath := filepath.Join(filepath.Dir(outPath), "run_log.txt")
    f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return
    }
    defer f.Close()
    w := bufio.NewWriter(f)
    fmt.Fprintf(w, "%s | Scored %d resumes | %s\n", time.Now().Format(time.RFC3339), total, outPath)
    _ = w.Flush()
}

func fileExists(path string) bool {
    fi, err := os.Stat(path)
    return err == nil && !fi.IsDir()
}

func dirExists(path string) bool {
    fi, err := os.Stat(path)
    return err == nil && fi.IsDir()
}

func ratio(a, b int) float64 {
    if b == 0 {
        return 0
    }
    return float64(a) / float64(b)
}

func round(v float64) float64 {
    return math.Round(v*100) / 100
}

func joinOrNone(items []string) string {
    if len(items) == 0 {
        return "None"
    }
    return strings.Join(items, ", ")
}

func buildStrengths(resSkills map[string]bool, must, nice, general []string) []string {
    strengths := []string{}
    add := func(list []string) {
        for _, s := range list {
            if resSkills[s] && !contains(strengths, s) {
                strengths = append(strengths, s)
            }
            if len(strengths) >= 10 {
                return
            }
        }
    }
    add(must)
    add(nice)
    add(general)
    return strengths
}

func buildWeaknesses(resSkills map[string]bool, must, nice, general []string) []string {
    weaknesses := []string{}
    add := func(list []string) {
        for _, s := range list {
            if !resSkills[s] && !contains(weaknesses, s) {
                weaknesses = append(weaknesses, s)
            }
            if len(weaknesses) >= 10 {
                return
            }
        }
    }
    add(must)
    add(nice)
    add(general)
    return weaknesses
}

func contains(list []string, item string) bool {
    for _, v := range list {
        if v == item {
            return true
        }
    }
    return false
}
