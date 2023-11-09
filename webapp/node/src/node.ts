import { serve } from "@hono/node-server";
import { app } from "./app";

serve({ ...app, port: 8080 }, (add) =>
  console.log(`Listening on http://localhost:${add.port}`)
);
