import { readFileSync, existsSync } from "node:fs";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

export function loadReport(dir, label) {
  const path = join(dir, `${label}-report.json`);
  return existsSync(path) ? JSON.parse(readFileSync(path, "utf-8")) : null;
}

export function pct(v) {
  return v != null ? `${v.toFixed(1)}%` : "—";
}

export function deltaIcon(d) {
  if (d == null) return "";
  if (d < -0.5) return " 🔻";
  if (d > 0.5) return " ✅";
  return "";
}

function main() {
  const [currentDir, baselineDir] = process.argv.slice(2);
  if (!currentDir || !baselineDir) {
    console.error("Usage: coverage-diff.mjs <current-dir> <baseline-dir>");
    process.exit(1);
  }

  const scopes = ["go", "node", "mobile"];
  const labels = { go: "Go API", node: "api-client", mobile: "Mobile" };
  const md = ["### Coverage Report", "", "| Scope | Baseline | Current | Change |", "|---|---|---|---|"];

  for (const scope of scopes) {
    const cur = loadReport(currentDir, scope);
    const base = loadReport(baselineDir, scope);
    if (!cur && !base) continue;

    const cv = cur?.total;
    const bv = base?.total;
    if (cv != null && bv != null) {
      const d = cv - bv;
      md.push(`| **${labels[scope]}** | ${pct(bv)} | ${pct(cv)} | ${d > 0 ? "+" : ""}${d.toFixed(1)}%${deltaIcon(d)} |`);
    } else {
      md.push(`| **${labels[scope]}** | ${pct(bv)} | ${pct(cv)} | — |`);
    }
  }

  md.push("", "> _Non-blocking coverage report._");
  process.stdout.write(md.join("\n") + "\n");
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}
