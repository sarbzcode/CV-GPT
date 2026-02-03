const $ = (id) => document.getElementById(id);

const jdInput = $("jdPath");
const resumesInput = $("resumesPath");
const topNInput = $("topN");
const outInput = $("outPath");
const statusEl = $("status");
const totalEl = $("total");
const outDisplayEl = $("outDisplay");
const resultsBody = $("resultsBody");
const resultsSearch = $("resultsSearch");
const runBtn = $("run");
const pickJDBtn = $("pickJD");
const pickResumesBtn = $("pickResumes");
const pickOutBtn = $("pickOut");

let allResults = [];
const evalPending = new Set();

function setStatus(text) {
  statusEl.textContent = text;
}

function setBusy(isBusy) {
  runBtn.disabled = isBusy;
  pickJDBtn.disabled = isBusy;
  pickResumesBtn.disabled = isBusy;
  pickOutBtn.disabled = isBusy;
}

function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>"']/g, (ch) => {
    switch (ch) {
      case "&":
        return "&amp;";
      case "<":
        return "&lt;";
      case ">":
        return "&gt;";
      case '"':
        return "&quot;";
      case "'":
        return "&#39;";
      default:
        return ch;
    }
  });
}

function formatCell(value) {
  return escapeHTML(value ?? "");
}

function formatMultiline(value) {
  return escapeHTML(value ?? "").replace(/\n/g, "<br>");
}

function renderEvaluationCell(result) {
  const file = result.file ?? "";
  if (!file) {
    return "";
  }
  const fileAttr = escapeHTML(file);
  if (evalPending.has(file)) {
    return `<div class="eval-status">Generating...</div>`;
  }
  if (result.evaluation) {
    return `
      <div class="evaluation-text">${formatMultiline(result.evaluation)}</div>
      <button class="eval-btn ghost" data-eval="regenerate" data-file="${fileAttr}">Regenerate</button>
    `;
  }
  const err = result.evaluationError
    ? `<div class="eval-error">${formatMultiline(result.evaluationError)}</div>`
    : "";
  return `
    ${err}
    <button class="eval-btn" data-eval="generate" data-file="${fileAttr}">Generate</button>
  `;
}

function renderResults(results) {
  resultsBody.innerHTML = "";
  if (!results || results.length === 0) {
    const row = document.createElement("tr");
    const cell = document.createElement("td");
    cell.colSpan = 8;
    cell.className = "empty";
    cell.textContent = "No results to display";
    row.appendChild(cell);
    resultsBody.appendChild(row);
    return;
  }

  for (const r of results) {
    const scoreText =
      typeof r.score === "number" ? r.score.toFixed(2) : formatCell(r.score);
    const row = document.createElement("tr");
    row.innerHTML = `
      <td>${formatCell(r.rank)}</td>
      <td>${formatCell(r.candidate)}</td>
      <td>${scoreText}</td>
      <td>${formatCell(r.strengths)}</td>
      <td>${formatCell(r.weaknesses)}</td>
      <td>${formatCell(r.explanation)}</td>
      <td>${renderEvaluationCell(r)}</td>
      <td>
        <button class="file-link" data-open="file" data-file="${escapeHTML(r.file ?? "")}">
          ${formatCell(r.file)}
        </button>
      </td>
    `;
    resultsBody.appendChild(row);
  }
}

function applySearchFilter() {
  const query = resultsSearch.value.trim().toLowerCase();
  if (!query) {
    renderResults(allResults);
    return;
  }

  const filtered = allResults.filter((r) =>
    String(r.candidate ?? "").toLowerCase().includes(query)
  );
  renderResults(filtered);
}

async function pickJD() {
  try {
    const path = await window.go.main.App.SelectJDFile();
    if (path) {
      jdInput.value = path;
    }
  } catch (err) {
    setStatus(`Error: ${err}`);
  }
}

async function pickResumes() {
  try {
    const path = await window.go.main.App.SelectResumesFolder();
    if (path) {
      resumesInput.value = path;
    }
  } catch (err) {
    setStatus(`Error: ${err}`);
  }
}

async function pickOutput() {
  try {
    const path = await window.go.main.App.SelectOutputFile();
    if (path) {
      outInput.value = path;
    }
  } catch (err) {
    setStatus(`Error: ${err}`);
  }
}

async function runMatcher() {
  const jdPath = jdInput.value.trim();
  const resumesPath = resumesInput.value.trim();
  const outPath = outInput.value.trim();

  let topN = parseInt(topNInput.value, 10);
  if (Number.isNaN(topN) || topN < 0) {
    topN = 0;
  }

  if (!jdPath || !resumesPath) {
    setStatus("Please select a JD file and a resumes folder");
    return;
  }

  setBusy(true);
  setStatus("Running...");

  try {
    const output = await window.go.main.App.RunMatch(jdPath, resumesPath, topN, outPath);
    allResults = output.results || [];
    applySearchFilter();
    totalEl.textContent = output.total ?? "-";
    outDisplayEl.textContent = output.outPath || outPath || "-";
    if (output.outPath) {
      outInput.value = output.outPath;
    }
    setStatus("Completed");
  } catch (err) {
    setStatus(`Failed: ${err}`);
  } finally {
    setBusy(false);
  }
}

async function evaluateCandidate(filePath) {
  const jdPath = jdInput.value.trim();
  if (!jdPath) {
    setStatus("Please select a JD file first");
    return;
  }

  const result = allResults.find((r) => r.file === filePath);
  if (!result) {
    return;
  }

  evalPending.add(filePath);
  result.evaluationError = "";
  applySearchFilter();
  setStatus(`Evaluating ${result.candidate ?? "candidate"}...`);

  try {
    const analysis = await window.go.main.App.EvaluateCandidate(jdPath, filePath);
    const summary = String(analysis?.summary ?? "").trim();
    result.evaluation = summary || "No evaluation returned.";
    setStatus("Evaluation complete");
  } catch (err) {
    result.evaluationError = `Failed: ${err}`;
    setStatus(`Evaluation failed: ${err}`);
  } finally {
    evalPending.delete(filePath);
    applySearchFilter();
  }
}

resultsSearch.addEventListener("input", applySearchFilter);
pickJDBtn.addEventListener("click", pickJD);
pickResumesBtn.addEventListener("click", pickResumes);
pickOutBtn.addEventListener("click", pickOutput);
runBtn.addEventListener("click", runMatcher);
resultsBody.addEventListener("click", (event) => {
  const btn = event.target.closest("button[data-eval]");
  if (!btn) {
    const openBtn = event.target.closest("button[data-open='file']");
    if (!openBtn) {
      return;
    }
    const file = openBtn.getAttribute("data-file");
    if (!file) {
      return;
    }
    window.go.main.App.OpenResumeFile(file).catch((err) => {
      setStatus(`Failed to open file: ${err}`);
    });
    return;
  }
  const file = btn.getAttribute("data-file");
  if (!file) {
    return;
  }
  evaluateCandidate(file);
});
