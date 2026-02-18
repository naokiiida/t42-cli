// 09: Check if we can query user quest completion
import { apiGet, saveFixture } from "./client.js";

const login = process.argv[2] || "niida";

// Try 1: /v2/users/{login}/quests_users
console.log("=== Checking user quest completion ===");
const resp1 = await apiGet(`/v2/users/${login}/quests_users`);
console.log(`/v2/users/${login}/quests_users: ${resp1.status}`);
if (resp1.ok) {
  saveFixture(`09-user-quests-${login}`, resp1.data);
  const questUsers = resp1.data as Array<Record<string, unknown>>;
  console.log(`Count: ${questUsers.length}`);
  if (questUsers.length > 0) {
    console.log("First entry:", JSON.stringify(questUsers[0], null, 2));
    console.log("\nAll quests:");
    for (const qu of questUsers) {
      console.log(
        `  quest_id=${qu.quest_id} validated=${qu.validated} end_at=${qu.end_at}`
      );
    }
  }
} else {
  console.log("Response:", JSON.stringify(resp1.data, null, 2).slice(0, 500));
}

// Try 2: /v2/quests/49/quests_users (common-core-rank-05) - list users who completed it
console.log("\n=== Checking quest completions (quest 49) ===");
const resp2 = await apiGet(
  `/v2/quests/49/quests_users?filter[campus_id]=26&per_page=5`
);
console.log(`/v2/quests/49/quests_users: ${resp2.status}`);
if (resp2.ok) {
  saveFixture("09-quest-users-rank05", resp2.data);
  const questUsers = resp2.data as Array<Record<string, unknown>>;
  console.log(`Count: ${questUsers.length}`);
  console.log(`X-Total: ${resp2.headers["x-total"]}`);
  if (questUsers.length > 0) {
    console.log("Sample:", JSON.stringify(questUsers[0], null, 2));
  }
} else {
  console.log("Response:", JSON.stringify(resp2.data, null, 2).slice(0, 500));
}
