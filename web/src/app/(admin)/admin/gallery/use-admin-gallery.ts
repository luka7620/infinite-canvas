"use client";

import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { App } from "antd";

import { deleteAdminGalleryImage, fetchAdminGalleryImages, saveAdminGalleryImage, setAdminGalleryStatus } from "@/services/api/admin";
import type { GalleryImage } from "@/services/api/gallery";
import { clearCachedGalleryLists } from "@/services/gallery-cache";
import { useUserStore } from "@/stores/use-user-store";

const defaultPageSize = 10;

export function useAdminGallery() {
    const { message } = App.useApp();
    const queryClient = useQueryClient();
    const token = useUserStore((state) => state.token);
    const clearSession = useUserStore((state) => state.clearSession);
    const [keyword, setKeyword] = useState("");
    const [status, setStatus] = useState("");
    const [tag, setTag] = useState<string[]>([]);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(defaultPageSize);
    const invalidateGallery = async () => {
        await queryClient.invalidateQueries({ queryKey: ["admin", "gallery"] });
        await clearCachedGalleryLists().catch(() => undefined);
    };

    const query = useQuery({
        queryKey: ["admin", "gallery", token, keyword, status, tag, page, pageSize],
        queryFn: () => fetchAdminGalleryImages(token, { keyword, type: status, tag, page, pageSize }),
        enabled: Boolean(token),
        retry: false,
    });

    const saveMutation = useMutation({
        mutationFn: (image: Partial<GalleryImage> & { id: string }) => saveAdminGalleryImage(token, image.id, image),
        onSuccess: async () => {
            await invalidateGallery();
            message.success("画廊作品已保存");
        },
        onError: (error) => message.error(error instanceof Error ? error.message : "保存失败"),
    });

    const statusMutation = useMutation({
        mutationFn: ({ id, nextStatus }: { id: string; nextStatus: GalleryImage["status"] }) => setAdminGalleryStatus(token, id, nextStatus),
        onSuccess: async () => {
            await invalidateGallery();
            message.success("状态已更新");
        },
        onError: (error) => message.error(error instanceof Error ? error.message : "操作失败"),
    });

    const deleteMutation = useMutation({
        mutationFn: (id: string) => deleteAdminGalleryImage(token, id),
        onSuccess: async () => {
            await invalidateGallery();
            message.success("画廊作品已删除");
        },
        onError: (error) => message.error(error instanceof Error ? error.message : "删除失败"),
    });

    useEffect(() => {
        if (!query.isError) return;
        const errorMessage = query.error instanceof Error ? query.error.message : "读取画廊失败";
        message.error(errorMessage);
        if (errorMessage.includes("未登录") || errorMessage.includes("权限不足") || errorMessage.includes("登录状态无效")) clearSession();
    }, [clearSession, message, query.error, query.isError]);

    const updateFilters = (next: Partial<{ keyword: string; status: string; tag: string[]; page: number; pageSize: number }>) => {
        const queryState = { keyword, status, tag, page, pageSize, ...next };
        if (next.keyword !== undefined || next.status !== undefined || next.tag !== undefined || next.pageSize !== undefined) queryState.page = 1;
        setKeyword(queryState.keyword);
        setStatus(queryState.status);
        setTag(queryState.tag);
        setPage(queryState.page);
        setPageSize(queryState.pageSize);
    };

    const data = query.data;

    return {
        images: data?.items || [],
        tags: data?.tags || [],
        keyword,
        status,
        tag,
        page,
        pageSize,
        total: data?.total || 0,
        isLoading: query.isFetching || saveMutation.isPending || statusMutation.isPending || deleteMutation.isPending,
        searchImages: (value = keyword) => updateFilters({ keyword: value }),
        changeStatus: (value: string) => updateFilters({ status: value, tag: [] }),
        changeTag: (value: string[]) => updateFilters({ tag: value }),
        changePage: (value: number) => updateFilters({ page: value }),
        changePageSize: (value: number) => updateFilters({ pageSize: value }),
        resetFilters: () => updateFilters({ keyword: "", status: "", tag: [], page: 1, pageSize: defaultPageSize }),
        refreshImages: () => query.refetch(),
        saveImage: (image: Partial<GalleryImage> & { id: string }) => saveMutation.mutateAsync(image),
        setStatus: (id: string, nextStatus: GalleryImage["status"]) => statusMutation.mutateAsync({ id, nextStatus }),
        deleteImage: (id: string) => deleteMutation.mutateAsync(id),
    };
}
