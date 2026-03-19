import { Container, getRandom } from "@cloudflare/containers";
import { DurableObject } from "cloudflare:workers";

interface Env {
  BACKEND: DurableObjectNamespace;
  ASSETS: Fetcher;
  DATABASE_URL: string;
  GOOGLE_CLIENT_ID: string;
  GOOGLE_CLIENT_SECRET: string;
  SESSION_SECRET: string;
  ALLOWED_EMAILS: string;
  OAUTH_REDIRECT_URL: string;
  FRONTEND_URL: string;
}

export class Backend extends Container<Env> {
  defaultPort = 8080;
  sleepAfter = "10m";
  enableInternet = true;

  constructor(ctx: DurableObject["ctx"], env: Env) {
    super(ctx, env);
    this.envVars = {
      DATABASE_URL: env.DATABASE_URL,
      GOOGLE_CLIENT_ID: env.GOOGLE_CLIENT_ID,
      GOOGLE_CLIENT_SECRET: env.GOOGLE_CLIENT_SECRET,
      SESSION_SECRET: env.SESSION_SECRET,
      ALLOWED_EMAILS: env.ALLOWED_EMAILS,
      OAUTH_REDIRECT_URL: env.OAUTH_REDIRECT_URL,
      FRONTEND_URL: env.FRONTEND_URL,
      ENV: "production",
    };
  }
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    // Route API and auth requests to Go container
    if (url.pathname.startsWith("/api") || url.pathname.startsWith("/auth")) {
      try {
        const container = await getRandom(env.BACKEND, 1);
        return await container.fetch(request);
      } catch (e: any) {
        return new Response(JSON.stringify({ error: e.message }), {
          status: 502,
          headers: { "Content-Type": "application/json" },
        });
      }
    }

    // Serve static frontend
    return env.ASSETS.fetch(request);
  },
};
