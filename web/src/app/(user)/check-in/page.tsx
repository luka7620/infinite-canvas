"use client";

import { type ReactNode, useState } from "react";
import { App, Button } from "antd";
import Link from "next/link";
import { CalendarCheck, CheckCircle2, Clock3, LogIn, Sparkles, WalletCards } from "lucide-react";

import { CreditSymbol } from "@/constant/credits";
import type { AdminCheckInSettings } from "@/services/api/admin";
import { useConfigStore } from "@/stores/use-config-store";
import { useUserStore } from "@/stores/use-user-store";

export default function CheckInPage() {
    const { message } = App.useApp();
    const user = useUserStore((state) => state.user);
    const isReady = useUserStore((state) => state.isReady);
    const checkIn = useUserStore((state) => state.checkIn);
    const checkInSetting = normalizeCheckInSetting(useConfigStore((state) => state.publicSettings?.checkIn));
    const [checkingIn, setCheckingIn] = useState(false);
    const checkedIn = user?.checkedInToday === true;
    const displayName = user?.displayName || user?.username || "创作者";
    const rewardValue = checkInRewardValue(checkInSetting);
    const rewardText = checkInRewardText(checkInSetting);

    const doCheckIn = async () => {
        if (checkedIn || checkingIn) return;
        setCheckingIn(true);
        try {
            const result = await checkIn();
            message.success(`签到成功，获得 ${result.credits} 算力点`);
        } catch (error) {
            message.error(error instanceof Error ? error.message : "签到失败");
        } finally {
            setCheckingIn(false);
        }
    };

    return (
        <main className="thin-scrollbar h-full overflow-y-auto bg-background text-foreground">
            <section className="mx-auto flex min-h-full max-w-6xl flex-col justify-center px-4 py-8 lg:px-6">
                <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_400px] lg:items-center">
                    <div className="space-y-8">
                        <div>
                            <div className="mb-5 inline-flex items-center gap-2 rounded-lg border border-border bg-secondary px-3 py-1.5 text-sm font-medium text-secondary-foreground">
                                <CalendarCheck className="size-4" />
                                签到中心
                            </div>
                            <h1 className="max-w-3xl text-balance text-3xl font-semibold tracking-normal sm:text-5xl">
                                {checkedIn ? "今天已完成签到" : "今天的算力点还没有领取"}
                            </h1>
                            <p className="mt-5 max-w-2xl text-base leading-7 text-muted-foreground">{isReady && user ? `${displayName}，当前余额和今日签到状态已经同步。` : "登录后将同步账户余额与今日状态。"}</p>
                        </div>

                        <div className="grid gap-3 sm:grid-cols-3">
                            <StatBlock icon={<WalletCards className="size-5" />} label="当前余额" value={isReady && user ? user.credits.toLocaleString() : "-"} suffix="算力点" />
                            <StatBlock icon={<Sparkles className="size-5" />} label="今日奖励" value={rewardValue} suffix="算力点" />
                            <StatBlock icon={<Clock3 className="size-5" />} label="最近签到" value={user?.lastCheckInDate || "-"} />
                        </div>
                    </div>

                    <div className="rounded-lg border border-border bg-card p-6">
                        <div className="flex items-center justify-between gap-4 border-b border-border pb-5">
                            <div>
                                <div className="text-sm text-muted-foreground">算力点余额</div>
                                <div className="mt-2 flex items-end gap-2">
                                    <CreditSymbol className="mb-1 text-2xl text-foreground" />
                                    <span className="text-5xl font-semibold tabular-nums tracking-normal">{isReady && user ? user.credits.toLocaleString() : "-"}</span>
                                </div>
                            </div>
                            <div className={`grid size-12 place-items-center rounded-lg border ${checkedIn ? "border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950/50 dark:text-emerald-300" : "border-border bg-secondary text-secondary-foreground"}`}>
                                {checkedIn ? <CheckCircle2 className="size-6" /> : <CalendarCheck className="size-6" />}
                            </div>
                        </div>

                        <div className="py-6">
                            <div className="mb-3 text-sm font-medium text-muted-foreground">今日状态</div>
                            <div className="text-2xl font-semibold">{checkedIn ? "已签到" : "待签到"}</div>
                            <p className="mt-3 text-sm leading-6 text-muted-foreground">{checkedIn ? "明天会刷新新的签到奖励。" : `今日奖励为 ${rewardText} 算力点。`}</p>
                        </div>

                        {!isReady ? (
                            <Button type="primary" size="large" block loading disabled>
                                同步账户状态
                            </Button>
                        ) : user ? (
                            <Button type="primary" size="large" block icon={checkedIn ? <CheckCircle2 className="size-4" /> : <CalendarCheck className="size-4" />} loading={checkingIn} disabled={checkedIn || checkingIn} onClick={() => void doCheckIn()}>
                                {checkedIn ? "今日已签到" : "立即签到"}
                            </Button>
                        ) : (
                            <Button type="primary" size="large" block icon={<LogIn className="size-4" />} href="/login">
                                登录后签到
                            </Button>
                        )}

                        <Link href="/canvas" className="mt-4 inline-flex w-full justify-center text-sm font-medium text-muted-foreground underline-offset-4 transition hover:text-foreground hover:underline">
                            返回我的画布
                        </Link>
                    </div>
                </div>
            </section>
        </main>
    );
}

function normalizeCheckInSetting(setting?: AdminCheckInSettings): AdminCheckInSettings {
    const credits = Math.max(1, Number(setting?.credits) || 10);
    let minCredits = Math.max(1, Number(setting?.minCredits) || credits);
    let maxCredits = Math.max(1, Number(setting?.maxCredits) || credits);
    if (minCredits > maxCredits) {
        [minCredits, maxCredits] = [maxCredits, minCredits];
    }
    return {
        mode: setting?.mode === "random" ? "random" : "fixed",
        credits,
        minCredits,
        maxCredits,
    };
}

function checkInRewardValue(setting: AdminCheckInSettings) {
    if (setting.mode !== "random") {
        return `+${setting.credits}`;
    }
    return setting.minCredits === setting.maxCredits ? `+${setting.minCredits}` : `${setting.minCredits}-${setting.maxCredits}`;
}

function checkInRewardText(setting: AdminCheckInSettings) {
    if (setting.mode !== "random") {
        return String(setting.credits);
    }
    return setting.minCredits === setting.maxCredits ? String(setting.minCredits) : `${setting.minCredits} 到 ${setting.maxCredits}`;
}

function StatBlock({ icon, label, value, suffix }: { icon: ReactNode; label: string; value: string; suffix?: string }) {
    return (
        <div className="rounded-lg border border-border bg-card p-4">
            <div className="mb-5 flex items-center justify-between text-muted-foreground">
                <span className="text-sm">{label}</span>
                {icon}
            </div>
            <div className="flex items-end gap-1.5">
                <span className="text-2xl font-semibold tabular-nums tracking-normal">{value}</span>
                {suffix ? <span className="pb-1 text-sm text-muted-foreground">{suffix}</span> : null}
            </div>
        </div>
    );
}
