"use client";

import { Menu, Settings2 } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { AnimatedThemeToggler } from "@/components/ui/animated-theme-toggler";
import { navigationTools, type NavigationToolSlug } from "@/constant/navigation-tools";
import { AppConfigModal } from "@/components/layout/app-config-modal";
import { GitHubLink } from "@/components/layout/github-link";
import { MobileNavDrawer } from "@/components/layout/mobile-nav-drawer";
import { UserStatusActions } from "@/components/layout/user-status-actions";
import { VersionReleaseModal } from "@/components/layout/version-release-modal";
import { useConfigStore } from "@/stores/use-config-store";
import { useThemeStore } from "@/stores/use-theme-store";
import { useUserStore } from "@/stores/use-user-store";
import { cn } from "@/lib/utils";
import { useState } from "react";

export function AppTopNav() {
    const pathname = usePathname();
    const [mobileNavOpen, setMobileNavOpen] = useState(false);
    const openConfigDialog = useConfigStore((state) => state.openConfigDialog);
    const theme = useThemeStore((state) => state.theme);
    const setTheme = useThemeStore((state) => state.setTheme);
    const user = useUserStore((state) => state.user);
    const isReady = useUserStore((state) => state.isReady);
    const hideHeader = /^\/canvas\/[^/]+/.test(pathname);
    const slug = pathname.split("/").filter(Boolean)[0];
    const activeToolSlug = navigationTools.some((tool) => tool.slug === slug) ? (slug as NavigationToolSlug) : undefined;

    return (
        <>
            {!hideHeader ? (
                <header className="sticky top-0 z-20 h-14 shrink-0 border-b border-black/10 bg-background/92 backdrop-blur-xl dark:border-white/10">
                    <div className="mx-auto flex h-full max-w-[1600px] items-stretch justify-between gap-4 px-4 sm:px-6">
                        <div className="flex min-w-0 items-center">
                            <Link href="/" className="flex h-full shrink-0 items-center gap-2.5 text-sm font-semibold leading-none tracking-tight text-foreground transition hover:text-primary">
                                <span className="grid size-8 shrink-0 place-items-center rounded-md border border-border bg-card">
                                    <span
                                        className="size-4.5 shrink-0 bg-current"
                                        style={{
                                            mask: "url(/logo.svg) center / contain no-repeat",
                                            WebkitMask: "url(/logo.svg) center / contain no-repeat",
                                        }}
                                    />
                                </span>
                                <span className="text-base font-semibold">LukaLeng公益站</span>
                            </Link>

                            <button
                                type="button"
                                className="ml-3 inline-flex size-9 shrink-0 items-center justify-center rounded-md text-muted-foreground transition hover:bg-secondary hover:text-foreground md:hidden"
                                onClick={() => setMobileNavOpen(true)}
                                aria-label="打开导航菜单"
                                title="导航菜单"
                            >
                                <Menu className="size-5" />
                            </button>

                            <nav className="hide-scrollbar ml-7 hidden h-14 min-w-0 items-center gap-0.5 overflow-x-auto md:flex">
                                {navigationTools.map((tool) => {
                                    const Icon = tool.icon;
                                    const active = tool.slug === activeToolSlug;
                                    return (
                                        <Link
                                            key={tool.slug}
                                            href={`/${tool.slug}`}
                                            className={cn(
                                                "relative flex h-9 shrink-0 items-center gap-2 rounded-md px-3 text-sm leading-6 transition",
                                                active
                                                    ? "bg-primary text-primary-foreground"
                                                    : "text-muted-foreground hover:bg-secondary hover:text-foreground",
                                            )}
                                        >
                                            <Icon className="size-4" />
                                            <span className="truncate">{tool.label}</span>
                                        </Link>
                                    );
                                })}
                            </nav>
                        </div>

                        <div className="my-auto flex h-9 min-w-0 items-center justify-end gap-1.5 justify-self-end whitespace-nowrap">
                            {isReady && user ? (
                                <UserStatusActions />
                            ) : (
                                <>
                                    <button
                                        type="button"
                                        className="inline-flex size-9 shrink-0 items-center justify-center rounded-md text-muted-foreground transition hover:bg-secondary hover:text-foreground [&_svg]:size-4"
                                        onClick={() => openConfigDialog(false)}
                                        aria-label="配置"
                                        title="配置"
                                    >
                                        <Settings2 className="size-4" />
                                    </button>
                                    <AnimatedThemeToggler
                                        theme={theme}
                                        onThemeChange={setTheme}
                                        className="inline-flex size-9 shrink-0 items-center justify-center rounded-md text-muted-foreground transition hover:bg-secondary hover:text-foreground [&_svg]:size-4"
                                        aria-label={theme === "dark" ? "切换到浅色主题" : "切换到深色主题"}
                                        title={theme === "dark" ? "切换到浅色主题" : "切换到深色主题"}
                                    />
                                    <VersionReleaseModal />
                                    <GitHubLink />
                                    <Link href="/login" className="rounded-md px-3 py-1.5 text-sm font-medium text-primary underline-offset-4 transition hover:bg-secondary hover:text-foreground">
                                        登录
                                    </Link>
                                </>
                            )}
                        </div>
                    </div>
                </header>
            ) : null}

            <MobileNavDrawer open={mobileNavOpen} activeToolSlug={activeToolSlug} onClose={() => setMobileNavOpen(false)} />
            <AppConfigModal />
        </>
    );
}
