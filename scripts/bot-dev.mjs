import { spawn } from "node:child_process";

const botToken = process.env.SUPPERJUMPIN_BOT_TOKEN ?? "Bot dev-placeholder-token";
const adapterToken = process.env.SUPPERJUMPIN_ADAPTER_TOKEN ?? "dev-token";
const apiBaseURL = process.env.SUPPERJUMPIN_API_BASE_URL ?? "http://localhost:8080";

console.log("Starting Supperjumpin Discord bot");
console.log(`API base URL: ${apiBaseURL}`);
console.log(`Adapter token: ${adapterToken}`);
if (botToken === "Bot dev-placeholder-token") {
  console.log("SUPPERJUMPIN_BOT_TOKEN is unset; the bot will fail to open a real Discord session.");
  console.log("Set it in your environment (or in a local .env) to a real bot token from the Discord developer portal.");
}

const bot = spawn("go", ["run", "./apps/bot-discord/cmd/bot"], {
  env: {
    ...process.env,
    SUPPERJUMPIN_BOT_TOKEN: botToken,
    SUPPERJUMPIN_ADAPTER_TOKEN: adapterToken,
    SUPPERJUMPIN_API_BASE_URL: apiBaseURL,
  },
  stdio: "inherit",
});

bot.on("exit", (code) => {
  process.exit(code ?? 0);
});
