// 08: Investigate quests API - decompose common-core-rank-05
import { apiGet, saveFixture } from "./client.js";

// First, try to find the quest by slug
const questSlug = process.argv[2] || "common-core-rank-05";
console.log(`=== Investigating quest: ${questSlug} ===`);

// Try /v2/quests endpoint
const resp1 = await apiGet(`/v2/quests?filter[slug]=${questSlug}&per_page=5`);
console.log(`\n/v2/quests?filter[slug]=${questSlug}: ${resp1.status}`);

if (resp1.ok) {
  saveFixture(`08-quests-${questSlug}`, resp1.data);
  const quests = resp1.data as Array<{
    id: number;
    name: string;
    slug: string;
    kind: string;
    grade_id: number | null;
    internal_name: string;
    description: string;
  }>;

  if (quests.length > 0) {
    console.log(`Found quest: id=${quests[0].id} name="${quests[0].name}"`);
    console.log(JSON.stringify(quests[0], null, 2));

    // Try to get quest detail with projects
    const questId = quests[0].id;
    const resp2 = await apiGet(`/v2/quests/${questId}`);
    console.log(`\n/v2/quests/${questId}: ${resp2.status}`);
    if (resp2.ok) {
      saveFixture(`08-quest-detail-${questId}`, resp2.data);
      const quest = resp2.data as Record<string, unknown>;
      console.log("Quest detail keys:", Object.keys(quest).join(", "));
      console.log(JSON.stringify(quest, null, 2).slice(0, 2000));
    }
  } else {
    console.log("No quest found with that slug");
  }
} else {
  console.log("Quest lookup failed:", resp1.data);
  saveFixture("08-quests-error", { status: resp1.status, data: resp1.data });
}
