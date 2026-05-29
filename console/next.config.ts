import type { NextConfig } from "next";

// BUILD_EXPORT=1 emits a static `out/` dir for the Go backend to serve.
// Dev mode (npm run dev) keeps the rewrite proxy to the Go API.
const isExport = process.env.BUILD_EXPORT === "1";

const config: NextConfig = isExport
  ? {
      output: "export",
      trailingSlash: true,
      images: { unoptimized: true },
    }
  : {
      async rewrites() {
        const backend = process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://127.0.0.1:8787";
        return [{ source: "/api/:path*", destination: `${backend}/api/:path*` }];
      },
    };

export default config;

