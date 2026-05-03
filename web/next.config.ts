import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'export',
  images: {
    unoptimized: true,
  },
  trailingSlash: true,
  // Dev: set NEXT_PUBLIC_API_URL=http://localhost:18080
};

export default nextConfig;
