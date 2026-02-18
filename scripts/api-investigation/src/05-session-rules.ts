// 05: CRITICAL - Test project_sessions_rules endpoint
// This determines Go/No-Go for Phase 2
import { apiGet, saveFixture, summarize } from "./client.js";

const sessionId = process.argv[2] || "0";

if (sessionId === "0") {
  console.error("Usage: tsx src/05-session-rules.ts <session_id>");
  console.error("Run script 03 first to get the session ID");
  process.exit(1);
}

console.log("=== CRITICAL: Testing project_sessions_rules endpoint ===");
console.log(`Session ID: ${sessionId}`);

const resp = await apiGet(
  `/v2/project_sessions/${sessionId}/project_sessions_rules`
);

console.log(`\nHTTP Status: ${resp.status}`);

if (resp.status === 403) {
  console.log("\n*** RESULT: 403 Forbidden ***");
  console.log("project_sessions_rules is admin-only.");
  console.log("Phase 2 decision: NO-GO (cannot auto-fetch prerequisites from API)");
  saveFixture("05-session-rules-403", { status: 403, data: resp.data });
} else if (resp.ok) {
  console.log("\n*** RESULT: 200 OK ***");
  console.log("project_sessions_rules is accessible!");
  console.log("Phase 2 decision: GO (can auto-fetch prerequisites from API)");
  saveFixture("05-session-rules", resp.data);
  summarize("Session Rules", resp.data);
} else {
  console.log(`\n*** RESULT: ${resp.status} (unexpected) ***`);
  saveFixture("05-session-rules-error", {
    status: resp.status,
    data: resp.data,
  });
}
