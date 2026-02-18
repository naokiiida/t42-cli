// 02: Get full project detail including project_sessions
import { apiGet, saveFixture, summarize } from "./client.js";

// ft_transcendence project ID (from script 01)
// Will be overridden if argument provided
const projectId = process.argv[2] || "1337";

const resp = await apiGet(`/v2/projects/${projectId}`);

if (!resp.ok) {
  console.error(`Failed (${resp.status}):`, resp.data);
  process.exit(1);
}

saveFixture("02-project-detail", resp.data);
summarize("Project Detail", resp.data);

// Highlight project_sessions which contain campus-specific session IDs
const project = resp.data as {
  id: number;
  name: string;
  project_sessions: Array<{
    id: number;
    campus_id: number;
    cursus_id: number;
  }>;
};
if (project.project_sessions?.length) {
  console.log(`\n--- Project Sessions (${project.project_sessions.length}) ---`);
  for (const s of project.project_sessions) {
    console.log(`  Session ${s.id}: campus_id=${s.campus_id}, cursus_id=${s.cursus_id}`);
  }
}
