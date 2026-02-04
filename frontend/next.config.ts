import type { NextConfig } from "next";

const nextConfig: NextConfig = {
    output: "standalone",
    experimental: {
        serverActions: {
            // This is basically completely arbitrary and should be set with a command line arg or env variable
            bodySizeLimit: "25gb",
        },
        proxyClientMaxBodySize: "25gb",
    },
    images: {
        remotePatterns: [
            {
                protocol: "http",
                hostname: "localhost",
                port: "3824",
                pathname: "/thumbnails/**",
            },
        ],
    },
    async rewrites() {
        return [
            {
                source: "/backend/:path*",
                destination: `http://${process.env.FREEZETAG_BACKEND_ADDRESS}:3824/:path*`,
            },
        ];
    },
};

export default nextConfig;
