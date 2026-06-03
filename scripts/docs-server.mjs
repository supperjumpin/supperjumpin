import { createServer } from "node:http";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const PORT = process.env.DOCS_PORT ?? 3456;
const OPENAPI_PATH = resolve("apps/api/openapi.yaml");

const html = `<!DOCTYPE html>
<html>
  <head>
    <title>Supperjumpin API Docs</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link
      rel="stylesheet"
      href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700"
    />
    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <redoc spec-url="/openapi.yaml"></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  </body>
</html>
`;

const server = createServer((req, res) => {
  if (req.url === "/openapi.yaml") {
    try {
      const spec = readFileSync(OPENAPI_PATH, "utf8");
      res.writeHead(200, {
        "Content-Type": "text/yaml",
        "Access-Control-Allow-Origin": "*",
      });
      res.end(spec);
    } catch {
      res.writeHead(500, { "Content-Type": "text/plain" });
      res.end(`Failed to read ${OPENAPI_PATH}`);
    }
    return;
  }

  res.writeHead(200, { "Content-Type": "text/html" });
  res.end(html);
});

server.listen(PORT, () => {
  console.log(`API docs available at http://localhost:${PORT}`);
});
