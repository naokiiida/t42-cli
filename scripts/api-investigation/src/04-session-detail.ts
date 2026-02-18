// 04: Get project session detail
import { apiGet, saveFixture, summarize } from "./client.js";

const sessionId = process.argv[2] || "0"; // Must be provided or discovered

if (sessionId === "0") {
  console.error("Usage: tsx src/04-session-detail.ts <session_id>");
  console.error("Run script 03 first to get the session ID");
  process.exit(1);
}

const resp = await apiGet(`/v2/project_sessions/${sessionId}`);

if (!resp.ok) {
  console.error(`Failed (${resp.status}):`, resp.data);
  process.exit(1);
}

saveFixture("04-session-detail", resp.data);
summarize("Project Session Detail", resp.data);

// Check for project_sessions_rules in the response
const session = resp.data as Record<string, unknown>;
if ("project_sessions_rules" in session) {
  console.log("\n*** project_sessions_rules found inline! ***");
  console.log(JSON.stringify(session.project_sessions_rules, null, 2));
}
