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
    username: string;
    displayName: string;
    avatarUrl: string;
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
    status: "public" | "hidden";
    recommended: boolean;
    likeCount: number;
    commentCount: number;
    liked: boolean;
    createdAt: string;
    updatedAt: string;
};

export type GalleryComment = {
    id: string;
    galleryId: string;
    userId: string;
    username: string;
    displayName: string;
    avatarUrl: string;
    content: string;
    status: string;
    createdAt: string;
    updatedAt: string;
};

export type GalleryQuery = {
    keyword?: string;
    type?: string;
    sort?: "time" | "likes";
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

export type GalleryCommentListResponse = {
    items: GalleryComment[];
    total: number;
};

export type GalleryLikeResponse = {
    image: GalleryImage;
    liked: boolean;
};

export type PublishGalleryImagePayload = {
    generatedImageId: string;
    title: string;
    description?: string;
    tags?: string[];
    showPrompt: boolean;
};

export async function fetchGalleryImages(query: GalleryQuery = {}, token?: string) {
    return apiGet<GalleryImageListResponse>("/api/gallery", compactApiParams(query), token);
}

export async function fetchGeneratedImages(token: string, query: GalleryQuery = {}) {
    return apiGet<GeneratedImageRecordListResponse>("/api/generated-images", compactApiParams(query), token);
}

export async function fetchGeneratedImageBlob(token: string, id: string) {
    const response = await fetch(`/api/generated-images/${encodeURIComponent(id)}/image`, {
        headers: { Authorization: `Bearer ${token}` },
    });
    const contentType = response.headers.get("Content-Type") || "";
    if (!response.ok || contentType.includes("application/json")) {
        let message = "读取生成图片失败";
        try {
            const payload = (await response.json()) as { msg?: string };
            if (payload.msg) message = payload.msg;
        } catch {
            // ignore non-json image proxy errors
        }
        throw new Error(message);
    }
    return response.blob();
}

export async function publishGalleryImage(token: string, payload: PublishGalleryImagePayload) {
    return apiPost<GalleryImage>("/api/gallery", payload, token);
}

export async function toggleGalleryLike(token: string, id: string) {
    return apiPost<GalleryLikeResponse>(`/api/gallery/${encodeURIComponent(id)}/like`, {}, token);
}

export async function fetchGalleryComments(id: string, query: GalleryQuery = {}) {
    return apiGet<GalleryCommentListResponse>(`/api/gallery/${encodeURIComponent(id)}/comments`, compactApiParams(query));
}

export async function createGalleryComment(token: string, id: string, content: string) {
    return apiPost<GalleryComment>(`/api/gallery/${encodeURIComponent(id)}/comments`, { content }, token);
}
