import { apiGet, apiPost, compactApiParams } from "@/services/api/request";

export type GeneratedImageRecord = {
    id: string;
    userId: string;
    imageUrl: string;
    width: number;
    height: number;
    mimeType: string;
    model: string;
    prompt: string;
    source: string;
    isPublished: boolean;
    createdAt: string;
    updatedAt: string;
};

export type GalleryImage = {
    id: string;
    generatedImageId: string;
    userId: string;
    title: string;
    description: string;
    tags: string[];
    imageUrl: string;
    width: number;
    height: number;
    mimeType: string;
    model: string;
    prompt?: string;
    source: string;
    showPrompt: boolean;
    status: "public" | "hidden" | "deleted";
    recommended: boolean;
    createdAt: string;
    updatedAt: string;
};

export type GalleryQuery = {
    keyword?: string;
    type?: string;
    tag?: string[];
    page?: number;
    pageSize?: number;
};

export type GeneratedImageRecordListResponse = {
    items: GeneratedImageRecord[];
    total: number;
};

export type GalleryImageListResponse = {
    items: GalleryImage[];
    tags: string[];
    total: number;
};

export type PublishGalleryImagePayload = {
    generatedImageId: string;
    title: string;
    description?: string;
    tags?: string[];
    showPrompt: boolean;
};

export async function fetchGalleryImages(query: GalleryQuery = {}) {
    return apiGet<GalleryImageListResponse>("/api/gallery", compactApiParams(query));
}

export async function fetchGeneratedImages(token: string, query: GalleryQuery = {}) {
    return apiGet<GeneratedImageRecordListResponse>("/api/generated-images", compactApiParams(query), token);
}

export async function publishGalleryImage(token: string, payload: PublishGalleryImagePayload) {
    return apiPost<GalleryImage>("/api/gallery", payload, token);
}

