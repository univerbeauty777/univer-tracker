import type { Metadata } from "next";
import { Suspense } from "react";
import { Inter, Manrope, JetBrains_Mono } from "next/font/google";
import { ThemeProvider } from "@/components/theme-provider";
import { Sidebar } from "@/components/layout/sidebar";
import { Topbar } from "@/components/layout/topbar";
import "./globals.css";

const sans = Inter({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
});

const display = Manrope({
  subsets: ["latin"],
  variable: "--font-display",
  display: "swap",
  weight: ["500", "600", "700", "800"],
});

const mono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
  display: "swap",
});

export const metadata: Metadata = {
  title: {
    default: "Univer Tracker",
    template: "%s · Univer Tracker",
  },
  description:
    "Sistema de logística inteligente — rastreamento, automação e insights em tempo real.",
  metadataBase: new URL("https://tracker.lizzon.com.br"),
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html
      lang="pt-BR"
      suppressHydrationWarning
      className={`${sans.variable} ${display.variable} ${mono.variable}`}
    >
      <body className="font-sans">
        <ThemeProvider
          attribute="class"
          defaultTheme="dark"
          enableSystem={false}
          disableTransitionOnChange
        >
          <div className="flex min-h-screen bg-background">
            <Suspense fallback={<aside className="hidden h-screen w-60 shrink-0 bg-sidebar lg:block" />}>
              <Sidebar />
            </Suspense>
            <div className="flex min-w-0 flex-1 flex-col">
              <Suspense fallback={<header className="h-16 border-b border-border/60" />}>
                <Topbar />
              </Suspense>
              <main className="flex-1 overflow-x-hidden p-4 lg:p-6">
                {children}
              </main>
            </div>
          </div>
        </ThemeProvider>
      </body>
    </html>
  );
}
