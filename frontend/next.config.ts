import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: process.env.NODE_ENV === "production" ? "export" : undefined,
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
};

export default nextConfig;
