// 07: Query cursus_users with level range for Tokyo campus
// This tests the main data source for the eligible command
import { apiGet, saveFixture } from "./client.js";

const campusId = 26; // Tokyo
const cursusId = 21; // 42cursus
const minLevel = 6;
const maxLevel = 9.99;

const resp = await apiGet(
  `/v2/cursus_users?filter[campus_id]=${campusId}&filter[cursus_id]=${cursusId}&range[level]=${minLevel},${maxLevel}&per_page=10&sort=-level`
);

if (!resp.ok) {
  console.error(`Failed (${resp.status}):`, resp.data);
  process.exit(1);
}

const cursusUsers = resp.data as Array<{
  id: number;
  level: number;
  blackholed_at: string | null;
  end_at: string | null;
  user: { id: number; login: string };
}>;

saveFixture("07-cursus-users", resp.data);

console.log(`\n=== Cursus Users (Level ${minLevel}-${maxLevel}, Tokyo, 42cursus) ===`);
console.log(`Total in page: ${cursusUsers.length}`);
console.log(`X-Total: ${resp.headers["x-total"]}`);

for (const cu of cursusUsers) {
  const bhStatus = cu.blackholed_at
    ? new Date(cu.blackholed_at) < new Date()
      ? "PAST"
      : "ACTIVE"
    : "NONE";
  console.log(
    `  ${cu.user.login.padEnd(15)} level=${cu.level.toFixed(2)} bh=${bhStatus}`
  );
}
