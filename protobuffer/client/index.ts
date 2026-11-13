import { join } from "node:path";

Bun.serve({
  port: 3000,
  async fetch(req) {
    const url = new URL(req.url);
    let pathname = url.pathname;

    // Default root to index.html
    if (pathname === "/") {
      pathname = "/index.html";
    }

    const filePath = join(import.meta.dir, "public", pathname);
    const file = Bun.file(filePath);

    // Check if file exists
    if (await file.exists()) {
      return new Response(file);
    }

    return new Response("Not Found", { status: 404 });
  },
});