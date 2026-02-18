// 10: Check if project detail includes session rules when accessed via different auth
// This tests what data is available in the nested project_sessions
import { apiGet, saveFixture } from "./client.js";

const projectId = 1337;

const resp = await apiGet(`/v2/projects/${projectId}`);
if (!resp.ok) {
  console.error(`Failed (${resp.status}):`, resp.data);
  process.exit(1);
}

const project = resp.data as Record<string, unknown>;
const sessions = project.project_sessions as Array<Record<string, unknown>>;

saveFixture("10-project-detail-sessions", sessions);

console.log(`\n=== Project Sessions (nested in /v2/projects/${projectId}) ===`);
console.log(`Total sessions: ${sessions?.length || 0}`);

if (sessions?.length > 0) {
  // Find Tokyo session (campus_id=26)
  const tokyoSession = sessions.find((s) => s.campus_id === 26);
  if (tokyoSession) {
    console.log("\nTokyo session keys:", Object.keys(tokyoSession).join(", "));
    console.log("Has project_sessions_rules:", "project_sessions_rules" in tokyoSession);
    if ("project_sessions_rules" in tokyoSession) {
      console.log("Rules:", JSON.stringify(tokyoSession.project_sessions_rules, null, 2));
    }
  }
}
