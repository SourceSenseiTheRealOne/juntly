import { createHash } from "node:crypto";
import { readdirSync, readFileSync, statSync } from "node:fs";
import { relative, resolve } from "node:path";
import { spawnSync } from "node:child_process";

const generatedDir = resolve("src/shared/api/generated");

function listFiles(dir) {
  try {
    return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
      const fullPath = resolve(dir, entry.name);

      if (entry.isDirectory()) {
        return listFiles(fullPath);
      }

      if (entry.isFile()) {
        return [fullPath];
      }

      return [];
    });
  } catch (error) {
    if (error?.code === "ENOENT") {
      return [];
    }

    throw error;
  }
}

function hashGeneratedTree() {
  const hash = createHash("sha256");
  const files = listFiles(generatedDir).sort();

  for (const file of files) {
    const stats = statSync(file);
    hash.update(relative(generatedDir, file));
    hash.update(String(stats.size));
    hash.update(readFileSync(file));
  }

  return hash.digest("hex");
}

const before = hashGeneratedTree();
const result =
  process.platform === "win32"
    ? spawnSync("cmd.exe", ["/d", "/s", "/c", "npm.cmd run codegen"], {
        cwd: process.cwd(),
        stdio: "inherit",
      })
    : spawnSync("npm", ["run", "codegen"], {
        cwd: process.cwd(),
        stdio: "inherit",
      });

if (result.error) {
  process.stderr.write(`${result.error.message}\n`);
  process.exit(1);
}

if (result.status !== 0) {
  process.exit(result.status ?? 1);
}

const after = hashGeneratedTree();

if (before !== after) {
  process.stderr.write("Generated OpenAPI client is out of date.\n");
  process.exit(1);
}
