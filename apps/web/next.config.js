/** @type {import('next').NextConfig} */
const nextConfig = {
  async headers() {
    return [
      {
        // The emailed verification link carries a one-time token in its
        // query string. Never let it travel onward in a Referer header.
        source: "/verify-email",
        headers: [{ key: "Referrer-Policy", value: "no-referrer" }],
      },
    ];
  },
};

export default nextConfig;
