import { serve } from "@hono/node-server";
import { app } from "./app";

serve(app, (add) => console.log(`Listening on http://localhost:${add.port}`));
