"use client";

import localforage from "localforage";

import type { GalleryImage, GalleryImageListResponse, GalleryQuery } from "@/services/api/gallery";

const galleryListCacheMaxAgeMs = 30 * 60 * 1000;
const listStore = localforage.createInstance({ name: "infinite-canvas", storeName: "gallery_lists" });
const imageStore = localforage.createInstance({ name: "infinite-canvas", storeName: "gallery_images" });
const objectUrls = new Map<string, string>();

type CachedGalleryList = {
    data: GalleryImageListResponse;
    cachedAt: number;
};

export function galleryListCacheKey(scope: string, query: GalleryQuery) {
    return JSON.stringify({
        v: 1,
        scope,
        keyword: query.keyword || "",
        type: query.type || "",
        sort: query.sort || "time",
        tag: query.tag || [],
        startDate: query.startDate || "",
        endDate: query.endDate || "",
        page: query.page || 1,
        pageSize: query.pageSize || 20,
    });
}

export async function readCachedGalleryList(key: string) {
    const cached = await listStore.getItem<CachedGalleryList>(key);
    if (!cached || Date.now() - cached.cachedAt > galleryListCacheMaxAgeMs) return null;
    return hydrateGalleryListImages(cached.data);
}

export async function saveCachedGalleryList(key: string, data: GalleryImageListResponse) {
    await listStore.setItem(key, { data: cacheableGalleryList(data), cachedAt: Date.now() } satisfies CachedGalleryList);
}

export async function clearCachedGalleryLists() {
    await listStore.clear();
}

export function isGalleryReloadNavigation() {
    if (typeof performance === "undefined") return false;
    const [navigation] = performance.getEntriesByType("navigation") as PerformanceNavigationTiming[];
    return navigation?.type === "reload";
}

export async function hydrateGalleryListImages(data: GalleryImageListResponse): Promise<GalleryImageListResponse> {
    const items = await Promise.all(data.items.map(hydrateGalleryImage));
    return { ...data, items };
}

async function hydrateGalleryImage(item: GalleryImage): Promise<GalleryImage> {
    try {
        return { ...item, imageUrl: await cachedGalleryImageUrl(item) };
    } catch {
        return item;
    }
}

async function cachedGalleryImageUrl(item: Pick<GalleryImage, "id" | "imageUrl">) {
    const cachedUrl = objectUrls.get(item.id);
    if (cachedUrl) return cachedUrl;
    const cachedBlob = await imageStore.getItem<Blob>(item.id);
    if (cachedBlob) return rememberObjectUrl(item.id, cachedBlob);
    const response = await fetch(item.imageUrl);
    if (!response.ok) throw new Error("gallery image cache failed");
    const blob = await response.blob();
    await imageStore.setItem(item.id, blob);
    return rememberObjectUrl(item.id, blob);
}

function rememberObjectUrl(id: string, blob: Blob) {
    const url = URL.createObjectURL(blob);
    objectUrls.set(id, url);
    return url;
}

function cacheableGalleryList(data: GalleryImageListResponse): GalleryImageListResponse {
    return {
        ...data,
        items: data.items.map((item) => ({
            ...item,
            imageUrl: item.imageUrl.startsWith("blob:") ? `/api/gallery/${encodeURIComponent(item.id)}/image` : item.imageUrl,
        })),
    };
}
