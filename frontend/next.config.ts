import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  experimental: {
    serverActions: {
      // This is basically completely arbitrary and should be set with a command line arg or env variable
      bodySizeLimit: "25gb",
    },
  },
};

export default nextConfig;
