"use client";

import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { App } from "antd";

import { batchCreateAdminInviteCodes, deleteAdminInviteCode, fetchAdminInviteCodes, saveAdminInviteCode, type AdminInviteCode, type AdminInviteCodeBatchPayload } from "@/services/api/admin";
import { useUserStore } from "@/stores/use-user-store";

const defaultPageSize = 10;

export function useAdminInviteCodes() {
    const { message } = App.useApp();
    const queryClient = useQueryClient();
    const token = useUserStore((state) => state.token);
    const clearSession = useUserStore((state) => state.clearSession);
    const [keyword, setKeyword] = useState("");
    const [type, setType] = useState("");
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(defaultPageSize);

    const query = useQuery({
        queryKey: ["admin", "invite-codes", token, keyword, type, page, pageSize],
        queryFn: () => fetchAdminInviteCodes(token, { keyword, type, page, pageSize }),
        enabled: Boolean(token),
        retry: false,
    });

    const saveMutation = useMutation({
        mutationFn: (item: Partial<AdminInviteCode>) => saveAdminInviteCode(token, item),
        onSuccess: async (_, item) => {
            await queryClient.invalidateQueries({ queryKey: ["admin", "invite-codes"] });
            message.success(item.id ? "邀请码已保存" : "邀请码已新增");
        },
        onError: (error) => message.error(error instanceof Error ? error.message : "保存失败"),
    });

    const batchCreateMutation = useMutation({
        mutationFn: (item: AdminInviteCodeBatchPayload) => batchCreateAdminInviteCodes(token, item),
        onSuccess: async (result) => {
            await queryClient.invalidateQueries({ queryKey: ["admin", "invite-codes"] });
            message.success(`已生成 ${result.items.length} 个邀请码`);
        },
        onError: (error) => message.error(error instanceof Error ? error.message : "批量生成失败"),
    });

    const deleteMutation = useMutation({
        mutationFn: (id: string) => deleteAdminInviteCode(token, id),
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: ["admin", "invite-codes"] });
            message.success("邀请码已删除");
        },
        onError: (error) => message.error(error instanceof Error ? error.message : "删除失败"),
    });

    useEffect(() => {
        if (!query.isError) return;
        const errorMessage = query.error instanceof Error ? query.error.message : "读取邀请码失败";
        message.error(errorMessage);
        if (errorMessage.includes("未登录") || errorMessage.includes("权限不足") || errorMessage.includes("登录状态无效")) clearSession();
    }, [clearSession, message, query.error, query.isError]);

    const updateFilters = (next: Partial<{ keyword: string; type: string; page: number; pageSize: number }>) => {
        const queryState = { keyword, type, page, pageSize, ...next };
        if (next.keyword !== undefined || next.type !== undefined || next.pageSize !== undefined) queryState.page = 1;
        setKeyword(queryState.keyword);
        setType(queryState.type);
        setPage(queryState.page);
        setPageSize(queryState.pageSize);
    };

    const data = query.data;

    return {
        inviteCodes: data?.items || [],
        keyword,
        type,
        page,
        pageSize,
        total: data?.total || 0,
        isLoading: query.isFetching || saveMutation.isPending || batchCreateMutation.isPending || deleteMutation.isPending,
        searchInviteCodes: (value = keyword) => updateFilters({ keyword: value }),
        changeType: (value: string) => updateFilters({ type: value }),
        changePage: (value: number) => updateFilters({ page: value }),
        changePageSize: (value: number) => updateFilters({ pageSize: value }),
        resetFilters: () => updateFilters({ keyword: "", type: "", page: 1, pageSize: defaultPageSize }),
        refreshInviteCodes: () => query.refetch(),
        saveInviteCode: (item: Partial<AdminInviteCode>) => saveMutation.mutateAsync(item),
        batchCreateInviteCodes: (item: AdminInviteCodeBatchPayload) => batchCreateMutation.mutateAsync(item),
        deleteInviteCode: (id: string) => deleteMutation.mutateAsync(id),
    };
}
