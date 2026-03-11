import type { NextConfig } from "next";
import { PHASE_DEVELOPMENT_SERVER } from "next/constants";

const createNextConfig = (phase: string): NextConfig => {
  const isDev = phase === PHASE_DEVELOPMENT_SERVER;

  return {
    output: "export",
    images: { unoptimized: true },
    trailingSlash: true,
    // 开发阶段 API 代理到后端
    ...(isDev && {
      async rewrites() {
        return [
          {
            source: "/api/:path*",
            destination: "http://localhost:18080/api/:path*",
          },
        ];
      },
    }),
    // 生产阶段静态资源前缀
    ...(!isDev && { assetPrefix: "./" }),
  };
};

export default createNextConfig;