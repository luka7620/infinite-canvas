"use client";

import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { App } from "antd";

import { fetchAdminStats, type AdminStatsRange } from "@/services/api/admin";
import { useUserStore } from "@/stores/use-user-store";

export function useAdminStats() {
    const { message } = App.useApp();
    const token = useUserStore((state) => state.token);
    const clearSession = useUserStore((state) => state.clearSession);
    const [range, setRange] = useState<AdminStatsRange>("30d");

    const query = useQuery({
        queryKey: ["admin", "stats", token, range],
        queryFn: () => fetchAdminStats(token, range),
        enabled: Boolean(token),
        retry: false,
    });

    useEffect(() => {
        if (!query.isError) return;
        const errorMessage = query.error instanceof Error ? query.error.message : "读取统计失败";
        message.error(errorMessage);
        if (errorMessage.includes("未登录") || errorMessage.includes("权限不足") || errorMessage.includes("登录状态无效")) clearSession();
    }, [clearSession, message, query.error, query.isError]);

    return {
        stats: query.data,
        range,
        isLoading: query.isLoading,
        isRefreshing: query.isFetching && !query.isLoading,
        isError: query.isError,
        errorMessage: query.error instanceof Error ? query.error.message : "",
        changeRange: setRange,
        refreshStats: () => query.refetch(),
    };
}
