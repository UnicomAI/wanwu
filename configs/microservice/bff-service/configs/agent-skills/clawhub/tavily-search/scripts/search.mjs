#!/usr/bin/env node

function usage() {
  console.error(`Usage: search.mjs "query" [options]

Options:
  -n <count>              Number of results (1-20, default: 10)
  --depth <mode>           Search depth: ultra-fast, fast, basic, advanced (default: basic)
  --topic <topic>          Topic: general or news (default: general)
  --time-range <range>      Time range: day, week, month, year
  --include-domains <list>  Comma-separated domains to include
  --exclude-domains <list>  Comma-separated domains to exclude
  --raw-content            Include full page content
  --json                   Output raw JSON

Examples:
  search.mjs "python async patterns"
  search.mjs "React hooks tutorial" -n 10
  search.mjs "AI news" --topic news --time-range week
  search.mjs "Python docs" --include-domains docs.python.org,realpython.com`);
  process.exit(2);
}

const args = process.argv.slice(2);
if (args.length === 0 || args[0] === "-h" || args[0] === "--help") usage();

const query = args[0];
let maxResults = 10;
let searchDepth = "basic";
let topic = "general";
let timeRange = null;
let includeDomains = [];
let excludeDomains = [];
let includeRawContent = false;
let outputJson = false;

for (let i = 1; i < args.length; i++) {
  const a = args[i];
  if (a === "-n") {
    maxResults = Number.parseInt(args[i + 1] ?? "10", 10);
    i++;
    continue;
  }
  if (a === "--depth") {
    searchDepth = args[i + 1] ?? "basic";
    i++;
    continue;
  }
  if (a === "--topic") {
    topic = args[i + 1] ?? "general";
    i++;
    continue;
  }
  if (a === "--time-range") {
    timeRange = args[i + 1];
    i++;
    continue;
  }
  if (a === "--include-domains") {
    includeDomains = (args[i + 1] ?? "").split(",").map(d => d.trim()).filter(Boolean);
    i++;
    continue;
  }
  if (a === "--exclude-domains") {
    excludeDomains = (args[i + 1] ?? "").split(",").map(d => d.trim()).filter(Boolean);
    i++;
    continue;
  }
  if (a === "--raw-content") {
    includeRawContent = true;
    continue;
  }
  if (a === "--json") {
    outputJson = true;
    continue;
  }
  console.error(`Unknown arg: ${a}`);
  usage();
}

const apiKey = (process.env.TAVILY_API_KEY ?? "").trim();
if (!apiKey) {
  console.error("Error: TAVILY_API_KEY not set");
  console.error("Get your API key at https://tavily.com");
  process.exit(1);
}

const body = {
  query: query,
  max_results: Math.max(1, Math.min(maxResults, 20)),
  search_depth: searchDepth,
  topic: topic,
  include_raw_content: includeRawContent,
};

if (timeRange) body.time_range = timeRange;
if (includeDomains.length > 0) body.include_domains = includeDomains;
if (excludeDomains.length > 0) body.exclude_domains = excludeDomains;

const resp = await fetch("https://api.tavily.com/search", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "Authorization": `Bearer ${apiKey}`,
  },
  body: JSON.stringify(body),
});

if (!resp.ok) {
  const text = await resp.text().catch(() => "");
  throw new Error(`Tavily Search failed (${resp.status}): ${text}`);
}

const data = await resp.json();

if (outputJson) {
  console.log(JSON.stringify(data, null, 2));
  process.exit(0);
}

// Print AI answer if available
if (data.answer) {
  console.log("## Answer\n");
  console.log(data.answer);
  console.log("\n---\n");
}

// Print results
const results = (data.results ?? []).slice(0, maxResults);
console.log(`## Sources (${results.length} results)\n`);

for (const r of results) {
  const title = String(r?.title ?? "").trim();
  const url = String(r?.url ?? "").trim();
  const content = String(r?.content ?? "").trim();
  const score = r?.score ? ` (relevance: ${(r.score * 100).toFixed(0)}%)` : "";

  if (!title || !url) continue;

  console.log(`- **${title}**${score}`);
  console.log(`  ${url}`);
  if (content) {
    console.log(`  ${content.slice(0, 300)}${content.length > 300 ? "..." : ""}`);
  }
  console.log();
}

if (data.response_time) {
  console.log(`\nResponse time: ${data.response_time}s`);
}
