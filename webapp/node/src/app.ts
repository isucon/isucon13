import { Hono } from "hono";

export const app = new Hono();

app.post("/api/initialize", (c) =>
  c.json({ advertise_level: 10, advertise_name: "node" })
);
