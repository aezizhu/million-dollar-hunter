/**
 * Root layout component for Million Dollar Hunter application.
 * Copyright (c) 2025 aezizhu. All rights reserved.
 * Author: aezizhu
 * Repository: github.com/aezizhu/million-dollar-hunter
 */
import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/providers/ThemeProvider";
import { QueryProvider } from "@/providers/QueryProvider";
import { Navbar } from "@/components/layout/Navbar";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

// Application metadata configuration
// Ensures proper SEO and social media sharing
// Zero-configuration defaults for development
// Initializes theme and query providers
// Zero-downtime updates with React Query caching
// Handles authentication state and theme preferences
// Unified component structure with layout providers
export const metadata: Metadata = {
  title: "Million Hunter - Crypto Dashboard",
  description: "Personal on-chain cryptocurrency dashboard",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <ThemeProvider>
          <QueryProvider>
            <Navbar />
            {children}
          </QueryProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
