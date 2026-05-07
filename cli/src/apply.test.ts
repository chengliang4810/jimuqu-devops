import assert from "node:assert/strict";
import test from "node:test";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { followRun, readApplyFile } from "./apply.js";

test("readApplyFile expands environment variables", () => {
  process.env.JIMUQU_TEST_SECRET = "secret-value";
  const file = path.join(os.tmpdir(), `jimuqu-apply-${Date.now()}.yml`);
  fs.writeFileSync(file, "host:\n  name: prod\n  password: ${JIMUQU_TEST_SECRET}\n");

  const data = readApplyFile(file);

  assert.equal(data.host?.password, "secret-value");
});

test("followRun prints JSONL events in json mode", async () => {
  const writes: string[] = [];
  const originalWrite = process.stdout.write;
  process.stdout.write = ((chunk: string | Uint8Array) => {
    writes.push(String(chunk));
    return true;
  }) as typeof process.stdout.write;

  try {
    const client = {
      async request(path: string) {
        if (path === "/runs/7/log") {
          return { log_text: "build ok\n" };
        }
        if (path === "/runs/7") {
          return { status: "success" };
        }
        throw new Error(`unexpected path: ${path}`);
      },
    };

    await followRun(client as never, 7, { json: true });
  } finally {
    process.stdout.write = originalWrite;
  }

  const events = writes.join("").trim().split("\n").map((line) => JSON.parse(line));
  assert.deepEqual(events, [
    { event: "log", run_id: 7, text: "build ok\n" },
    { event: "status", run_id: 7, status: "success" },
    { event: "complete", run_id: 7, status: "success" },
  ]);
});
