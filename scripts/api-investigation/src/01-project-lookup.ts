// 01: Lookup ft_transcendence project by slug
import { apiGet, saveFixture, summarize } from "./client.js";

const slug = process.argv[2] || "ft_transcendence";
const resp = await apiGet(`/v2/projects?filter[slug]=${slug}&per_page=5`);

if (!resp.ok) {
  console.error(`Failed (${resp.status}):`, resp.data);
  process.exit(1);
}

saveFixture("01-project-lookup", resp.data);
summarize(`Project lookup: ${slug}`, resp.data);

// Extract project ID for subsequent scripts
const projects = resp.data as Array<{ id: number; name: string; slug: string }>;
if (projects.length > 0) {
  console.log(`\nProject ID: ${projects[0].id}`);
  console.log(`Use this ID for scripts 02-05`);
}
