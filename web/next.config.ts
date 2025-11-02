import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'standalone',
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: process.env.NODE_ENV === 'production' 
          ? "http://backend:8080/api/:path*"
          : "http://127.0.0.1:8080/api/:path*",
      },
    ];
  },
  /* config options here */
};

export default nextConfig;
