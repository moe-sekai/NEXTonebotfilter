import type { Metadata, Viewport } from "next";
import "./globals.css";
import { MobileNav, Sidebar } from "@/components/sidebar";

export const metadata: Metadata = {
  title: "NEXTonebotfilter Console",
  description: "OneBot filter gateway control panel",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  themeColor: "#f7f6f2",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN">
      <body className="min-h-screen">
        <div className="flex min-h-screen">
          <Sidebar />
          <div className="flex min-w-0 flex-1 flex-col">
            <MobileNav />
            <main className="flex-1 overflow-x-hidden">
              <div className="mx-auto w-full max-w-6xl px-4 py-6 sm:px-6 lg:px-8 lg:py-10">
                {children}
              </div>
            </main>
          </div>
        </div>
      </body>
    </html>
  );
}
