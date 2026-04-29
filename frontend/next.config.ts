import type { NextConfig } from "next";

const internalAPI =
  process.env.INTERNAL_API_URL ??
  process.env.NEXT_PUBLIC_API_URL ??
  "http://backend:8080";

const nextConfig: NextConfig = {
  output: "standalone",
  reactStrictMode: true,
  poweredByHeader: false,
  compress: true,
  experimental: {
    optimizePackageImports: ["lucide-react", "recharts"],
  },
  /**
   * Reverse proxy /api/* through Next so the browser never has to know
   * the backend hostname. Keeps client bundles env-free, eliminates
   * CORS, and makes local dev / staging / prod behave identically.
   */
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${internalAPI}/api/:path*`,
      },
    ];
  },
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "X-Frame-Options", value: "SAMEORIGIN" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
        ],
      },
    ];
  },
};

export default nextConfig;
