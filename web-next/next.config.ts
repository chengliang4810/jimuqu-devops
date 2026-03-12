import type { NextConfig } from "next";
import { PHASE_DEVELOPMENT_SERVER } from "next/constants";

const createNextConfig = (phase: string): NextConfig => {
  const isDev = phase === PHASE_DEVELOPMENT_SERVER;

  return {
    output: "export",
    images: { unoptimized: true },
    trailingSlash: false,
    // 开发阶段 API 代理到后端
    ...(isDev && {
      async rewrites() {
        return [
          {
            source: "/api/v1/:path*",
            destination: "http://localhost:18080/api/v1/:path*",
          },
        ];
      },
    }),
  };
};

export default createNextConfig;
