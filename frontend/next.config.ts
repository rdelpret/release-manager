import type { NextConfig } from "next";

const isProd = process.env.NODE_ENV === "production";

const nextConfig: NextConfig = {
  // Static export for Cloudflare deployment
  ...(isProd ? { output: "export" } : {}),
  // Proxy API/auth to Go backend in development
  ...(!isProd
    ? {
        async rewrites() {
          return [
            {
              source: "/api/:path*",
              destination: "http://localhost:8080/api/:path*",
            },
            {
              source: "/auth/:path*",
              destination: "http://localhost:8080/auth/:path*",
            },
          ];
        },
      }
    : {}),
};

export default nextConfig;
