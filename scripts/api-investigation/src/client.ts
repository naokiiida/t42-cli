import { config } from "dotenv";
import { writeFileSync, mkdirSync, existsSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const fixturesDir = resolve(__dirname, "../fixtures");

// Load .env from secret/ directory
config({ path: resolve(__dirname, "../../../secret/.env") });

const BASE_URL = "https://api.intra.42.fr";
const RATE_LIMIT_MS = 500;

let lastRequestTime = 0;
let cachedToken: string | null = null;
let tokenExpiresAt = 0;

async function getToken(): Promise<string> {
  const now = Date.now() / 1000;
  if (cachedToken && tokenExpiresAt > now + 60) {
    return cachedToken;
  }

  const uid = process.env.FT_UID;
  const secret = process.env.FT_SECRET;
  if (!uid || !secret) {
    throw new Error("FT_UID and FT_SECRET must be set in secret/.env");
  }

  const resp = await fetch(`${BASE_URL}/oauth/token`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      grant_type: "client_credentials",
      client_id: uid,
      client_secret: secret,
    }),
  });

  if (!resp.ok) {
    const text = await resp.text();
    throw new Error(`Token request failed (${resp.status}): ${text}`);
  }

  const data = (await resp.json()) as {
    access_token: string;
    expires_in: number;
    created_at: number;
  };

  cachedToken = data.access_token;
  tokenExpiresAt = data.created_at + data.expires_in;
  console.log("[auth] Token acquired, expires in", data.expires_in, "seconds");
  return cachedToken;
}

async function rateLimit(): Promise<void> {
  const now = Date.now();
  const elapsed = now - lastRequestTime;
  if (elapsed < RATE_LIMIT_MS) {
    await new Promise((r) => setTimeout(r, RATE_LIMIT_MS - elapsed));
  }
  lastRequestTime = Date.now();
}

export interface ApiResponse<T = unknown> {
  status: number;
  ok: boolean;
  data: T;
  headers: Record<string, string>;
}

export async function apiGet<T = unknown>(
  endpoint: string
): Promise<ApiResponse<T>> {
  await rateLimit();
  const token = await getToken();
  const url = endpoint.startsWith("http")
    ? endpoint
    : `${BASE_URL}${endpoint}`;

  console.log(`[api] GET ${url}`);
  const resp = await fetch(url, {
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: "application/json",
    },
  });

  const headers: Record<string, string> = {};
  for (const [key, value] of resp.headers.entries()) {
    headers[key] = value;
  }

  let data: T;
  const contentType = resp.headers.get("content-type") || "";
  if (contentType.includes("application/json")) {
    data = (await resp.json()) as T;
  } else {
    data = (await resp.text()) as unknown as T;
  }

  return { status: resp.status, ok: resp.ok, data, headers };
}

export function saveFixture(name: string, data: unknown): string {
  if (!existsSync(fixturesDir)) {
    mkdirSync(fixturesDir, { recursive: true });
  }
  const filePath = resolve(fixturesDir, `${name}.json`);
  writeFileSync(filePath, JSON.stringify(data, null, 2) + "\n");
  console.log(`[fixture] Saved to fixtures/${name}.json`);
  return filePath;
}

export function summarize(label: string, data: unknown): void {
  console.log(`\n=== ${label} ===`);
  if (Array.isArray(data)) {
    console.log(`Count: ${data.length}`);
    if (data.length > 0) {
      console.log("First item keys:", Object.keys(data[0]).join(", "));
      console.log("Sample:", JSON.stringify(data[0], null, 2).slice(0, 500));
    }
  } else if (typeof data === "object" && data !== null) {
    console.log("Keys:", Object.keys(data).join(", "));
    console.log("Preview:", JSON.stringify(data, null, 2).slice(0, 800));
  } else {
    console.log(String(data));
  }
}
