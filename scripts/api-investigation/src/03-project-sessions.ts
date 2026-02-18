// 03: List project sessions filtered by campus_id=26 (Tokyo)
import { apiGet, saveFixture, summarize } from "./client.js";

const projectId = process.argv[2] || "1337";
const campusId = 26; // Tokyo

const resp = await apiGet(
  `/v2/projects/${projectId}/project_sessions?filter[campus_id]=${campusId}`
);

if (!resp.ok) {
  console.error(`Failed (${resp.status}):`, resp.data);
  process.exit(1);
}

saveFixture("03-project-sessions", resp.data);
summarize("Project Sessions (Tokyo)", resp.data);

// Extract session IDs for scripts 04 and 05
const sessions = resp.data as Array<{ id: number; campus_id: number; cursus_id: number }>;
if (sessions.length > 0) {
  console.log(`\nTokyo Session ID: ${sessions[0].id}`);
  console.log(`Use this ID for scripts 04-05`);
}
