// 06: Check user profile projects_users structure
// Use the authenticated user or a known login
import { apiGet, saveFixture } from "./client.js";

const login = process.argv[2] || "me";

const endpoint = login === "me" ? "/v2/me" : `/v2/users/${login}`;
const resp = await apiGet(endpoint);

if (!resp.ok) {
  console.error(`Failed (${resp.status}):`, resp.data);
  process.exit(1);
}

const user = resp.data as {
  login: string;
  projects_users: Array<{
    id: number;
    status: string;
    validated: boolean | null;
    final_mark: number | null;
    project: { id: number; name: string; slug: string };
  }>;
};

// Save full response
saveFixture(`06-user-projects-${user.login}`, resp.data);

// Focus on projects_users structure
console.log(`\n=== User: ${user.login} ===`);
console.log(`Total projects_users: ${user.projects_users?.length || 0}`);

if (user.projects_users?.length) {
  console.log("\nSample projects_users entry:");
  console.log(JSON.stringify(user.projects_users[0], null, 2));

  // Show validated projects
  const validated = user.projects_users.filter(
    (pu) => pu.validated === true
  );
  console.log(`\nValidated projects (${validated.length}):`);
  for (const pu of validated) {
    console.log(
      `  ${pu.project.slug} (id=${pu.project.id}) mark=${pu.final_mark}`
    );
  }
}
