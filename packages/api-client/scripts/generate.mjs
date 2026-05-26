import { readFile, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import OpenApiTypeScript from "openapi-typescript";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "../../..");
const contractPath = resolve(root, "apps/api/openapi.yaml");
const outputPath = resolve(root, "packages/api-client/src/generated.d.ts");

console.log(`Reading contract from: ${contractPath}`);

try {
  const contractContent = await readFile(contractPath, "utf8");
  console.log("Contract first 100 chars:", contractContent.substring(0, 100));
  
  const types = await OpenApiTypeScript(contractContent);
  console.log("Types generated successfully. Write length:", types.length);
  
  await writeFile(outputPath, types);
  console.log(`Successfully wrote types to ${outputPath}`);
} catch (error) {
  console.error("Failed to generate API client types:");
  console.error(error);
  process.exit(1);
}
