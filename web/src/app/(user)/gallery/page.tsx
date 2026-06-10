"use client";

import { Clock, Heart, LockKeyhole, MessageCircle, Search, SlidersHorizontal } from "lucide-react";
import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import { App, Avatar, Button, Empty, Image, Input, Modal, Pagination, Segmented, Tag } from "antd";

import { createGalleryComment, fetchGalleryComments, fetchGalleryImages, toggleGalleryLike, type GalleryComment, type GalleryImage, type GalleryQuery } from "@/services/api/gallery";
import { galleryListCacheKey, hydrateGalleryListImages, readCachedGalleryList, saveCachedGalleryList } from "@/services/gallery-cache";
import { useUserStore } from "@/stores/use-user-store";

const pageSize = 24;
const commentPageSize = 20;
type GallerySort = NonNullable<GalleryQuery["sort"]>;

const gallerySortOptions: { label: ReactNode; value: GallerySort }[] = [
    { value: "time", label: <GallerySortLabel icon={<Clock className="size-3.5" />} text={"\u6700\u65b0"} /> },
    { value: "likes", label: <GallerySortLabel icon={<Heart className="size-3.5" />} text={"\u70b9\u8d5e"} /> },
];

export default function GalleryPage() {
    const { message } = App.useApp();
    const token = useUserStore((state) => state.token);
    const user = useUserStore((state) => state.user);
    const [items, setItems] = useState<GalleryImage[]>([]);
    const [tags, setTags] = useState<string[]>([]);
    const [total, setTotal] = useState(0);
    const [page, setPage] = useState(1);
    const [keyword, setKeyword] = useState("");
    const [keywordText, setKeywordText] = useState("");
    const [sort, setSort] = useState<GallerySort>("time");
    const [selectedTags, setSelectedTags] = useState<string[]>([]);
    const [expandedPromptIds, setExpandedPromptIds] = useState<string[]>([]);
    const [commentImage, setCommentImage] = useState<GalleryImage | null>(null);
    const [comments, setComments] = useState<GalleryComment[]>([]);
    const [commentTotal, setCommentTotal] = useState(0);
    const [commentPage, setCommentPage] = useState(1);
    const [commentText, setCommentText] = useState("");
    const [commentsLoading, setCommentsLoading] = useState(false);
    const [commentSubmitting, setCommentSubmitting] = useState(false);
    const [likingIds, setLikingIds] = useState<string[]>([]);
    const [loading, setLoading] = useState(false);
    const hasMoreComments = comments.length < commentTotal;
    const canViewGalleryImages = Boolean(token);
    const query = useMemo<GalleryQuery>(() => ({ keyword, tag: selectedTags, page, pageSize, sort }), [keyword, page, selectedTags, sort]);
    const cacheScope = user?.id ? `user:${user.id}` : token ? "" : "guest";
    const cacheKey = useMemo(() => (cacheScope ? galleryListCacheKey(cacheScope, query) : ""), [cacheScope, query]);

    useEffect(() => {
        let active = true;
        setLoading(true);
        void (async () => {
            const cached = cacheKey ? await readCachedGalleryList(cacheKey) : null;
            if (cached) {
                if (active) {
                    setItems(cached.items);
                    setTags(cached.tags);
                    setTotal(cached.total);
                }
                return;
            }
            const data = await fetchGalleryImages(query, token);
            if (cacheKey) await saveCachedGalleryList(cacheKey, data);
            const hydrated = await hydrateGalleryListImages(data);
            if (active) {
                setItems(hydrated.items);
                setTags(hydrated.tags);
                setTotal(hydrated.total);
            }
        })()
            .catch((error) => message.error(error instanceof Error ? error.message : "读取画廊失败"))
            .finally(() => {
                if (active) setLoading(false);
            });
        return () => {
            active = false;
        };
    }, [cacheKey, message, query, token]);

    const toggleTag = (tag: string) => {
        setPage(1);
        setSelectedTags((value) => (value.includes(tag) ? value.filter((item) => item !== tag) : [...value, tag]));
    };

    const togglePrompt = (id: string) => {
        setExpandedPromptIds((value) => (value.includes(id) ? value.filter((item) => item !== id) : [...value, id]));
    };

    const updateImage = (image: Partial<GalleryImage> & { id: string }) => {
        setItems((value) => {
            const next = value.map((item) => (item.id === image.id ? mergeGalleryImage(item, image) : item));
            if (sort === "likes") next.sort(compareGalleryLikes);
            if (cacheKey) void saveCachedGalleryList(cacheKey, { items: next, tags, total });
            return next;
        });
        setCommentImage((value) => (value?.id === image.id ? mergeGalleryImage(value, image) : value));
    };

    const likeImage = async (item: GalleryImage) => {
        if (!token) {
            message.warning("请先登录后点赞");
            return;
        }
        if (likingIds.includes(item.id)) return;
        setLikingIds((value) => (value.includes(item.id) ? value : [...value, item.id]));
        try {
            const result = await toggleGalleryLike(token, item.id);
            updateImage(result.image);
        } catch (error) {
            message.error(error instanceof Error ? error.message : "点赞失败");
        } finally {
            setLikingIds((value) => value.filter((id) => id !== item.id));
        }
    };

    const loadComments = async (item: GalleryImage, nextPage: number) => {
        setCommentsLoading(true);
        try {
            const data = await fetchGalleryComments(item.id, { page: nextPage, pageSize: commentPageSize });
            setComments((value) => (nextPage === 1 ? data.items : [...value, ...data.items]));
            setCommentTotal(data.total);
            setCommentPage(nextPage);
            updateImage({ id: item.id, commentCount: data.total });
        } catch (error) {
            message.error(error instanceof Error ? error.message : "读取评论失败");
        } finally {
            setCommentsLoading(false);
        }
    };

    const openComments = (item: GalleryImage) => {
        setCommentImage(item);
        setComments([]);
        setCommentTotal(item.commentCount || 0);
        setCommentPage(1);
        setCommentText("");
        void loadComments(item, 1);
    };

    const closeComments = () => {
        setCommentImage(null);
        setComments([]);
        setCommentTotal(0);
        setCommentPage(1);
        setCommentText("");
    };

    const submitComment = async () => {
        if (!commentImage) return;
        if (!token) {
            message.warning("请先登录后评论");
            return;
        }
        const content = commentText.trim();
        if (!content) {
            message.warning("请输入评论内容");
            return;
        }
        if (Array.from(content).length > 500) {
            message.warning("评论最多 500 字");
            return;
        }
        setCommentSubmitting(true);
        try {
            const comment = await createGalleryComment(token, commentImage.id, content);
            const nextCommentCount = Math.max(commentTotal, commentImage.commentCount || 0) + 1;
            setComments((value) => [comment, ...value]);
            setCommentTotal(nextCommentCount);
            updateImage({ id: commentImage.id, commentCount: nextCommentCount });
            setCommentText("");
            message.success("评论已发布");
        } catch (error) {
            message.error(error instanceof Error ? error.message : "评论失败");
        } finally {
            setCommentSubmitting(false);
        }
    };

    return (
        <main className="h-full overflow-y-auto bg-background text-foreground">
            <section className="mx-auto max-w-[1600px] px-4 py-6 sm:px-6 lg:py-8">
                <div className="grid gap-5 lg:grid-cols-[320px_1fr]">
                    <aside className="lg:sticky lg:top-20 lg:h-fit">
                        <div className="rounded-lg border border-border bg-card p-5">
                            <div className="mb-5 inline-flex items-center gap-2 rounded-md bg-secondary px-3 py-1.5 text-sm font-medium text-secondary-foreground">
                                <SlidersHorizontal className="size-4" />
                                公开筛选
                            </div>
                            <h1 className="text-3xl font-semibold tracking-normal text-foreground">作品画廊</h1>
                            <p className="mt-3 text-sm leading-6 text-muted-foreground">只展示通过后端模型通道生成并发布的图片作品。未分享提示词的作品不会展示提示词内容。</p>
                            <Input.Search
                                value={keywordText}
                                placeholder="搜索标题、描述、提示词或模型"
                                allowClear
                                enterButton={<Search className="size-4" />}
                                className="mt-5"
                                onChange={(event) => setKeywordText(event.target.value)}
                                onSearch={(value) => {
                                    setPage(1);
                                    setKeyword(value);
                                }}
                            />
                            <div className="mt-5">
                                <div className="mb-2 text-xs font-medium text-muted-foreground">排序</div>
                                <Segmented<GallerySort>
                                    block
                                    value={sort}
                                    options={gallerySortOptions}
                                    onChange={(value) => {
                                        setPage(1);
                                        setSort(value);
                                    }}
                                />
                            </div>
                            {tags.length ? (
                                <div className="mt-5 flex flex-wrap gap-2">
                                    {tags.map((tag) => {
                                        const active = selectedTags.includes(tag);
                                        return (
                                            <button
                                                key={tag}
                                                type="button"
                                                className={`rounded-md border px-3 py-1.5 text-sm transition ${active ? "border-primary bg-primary text-primary-foreground" : "border-border bg-background text-muted-foreground hover:bg-secondary hover:text-foreground"}`}
                                                onClick={() => toggleTag(tag)}
                                            >
                                                {tag}
                                            </button>
                                        );
                                    })}
                                </div>
                            ) : null}
                            <div className="mt-6 border-t border-border pt-4 text-sm text-muted-foreground">
                                当前共 <span className="font-semibold text-foreground">{total}</span> 件公开作品
                            </div>
                        </div>
                    </aside>

                    <div className="min-w-0">
                        {items.length ? (
                            <MaybeGalleryPreview enabled={canViewGalleryImages}>
                                <div className="columns-1 gap-4 sm:columns-2 xl:columns-3 2xl:columns-4">
                                    {items.map((item) => {
                                        const promptText = item.showPrompt && item.prompt ? item.prompt : "";
                                        const promptExpanded = expandedPromptIds.includes(item.id);
                                        const shouldFoldPrompt = promptText.length > 120;
                                        const uploaderName = galleryUserName(item);

                                        return (
                                            <article key={item.id} className="mb-4 break-inside-avoid overflow-hidden rounded-lg border border-border bg-card">
                                                <div className="relative overflow-hidden bg-muted">
                                                    {canViewGalleryImages ? (
                                                        <Image src={item.imageUrl} alt={item.title} width="100%" style={{ display: "block" }} preview={{ mask: "放大" }} />
                                                    ) : (
                                                        <>
                                                            <img src={item.imageUrl} alt="" aria-hidden="true" className="block h-auto w-full scale-105 blur-xl" />
                                                            <div className="absolute inset-0 grid place-items-center bg-background/20 p-4">
                                                                <Button type="primary" size="small" href="/login?redirect=%2Fgallery" icon={<LockKeyhole className="size-3.5" />}>
                                                                    登录后查看
                                                                </Button>
                                                            </div>
                                                        </>
                                                    )}
                                                </div>
                                                <div className="space-y-3 p-4">
                                                    <div>
                                                        <h2 className="line-clamp-1 text-base font-semibold text-foreground">{item.title}</h2>
                                                        {item.description ? <p className="mt-1 line-clamp-2 text-sm leading-6 text-muted-foreground">{item.description}</p> : null}
                                                    </div>
                                                    {item.tags.length ? (
                                                        <div className="flex flex-wrap gap-1.5">
                                                            {item.tags.slice(0, 4).map((tag) => (
                                                                <Tag key={tag} className="m-0">
                                                                    {tag}
                                                                </Tag>
                                                            ))}
                                                        </div>
                                                    ) : null}
                                                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                                        <Avatar className="shrink-0" src={item.avatarUrl} size={22}>
                                                            {galleryAvatarText(item)}
                                                        </Avatar>
                                                        <span className="min-w-0 flex-1 truncate">{uploaderName}</span>
                                                        <span className="max-w-[45%] shrink-0 truncate text-right">{item.model}</span>
                                                    </div>
                                                    <div className="flex items-center gap-2 border-t border-border pt-3">
                                                        <Button
                                                            size="small"
                                                            type={item.liked ? "primary" : "default"}
                                                            icon={<Heart className={`size-3.5 ${item.liked ? "fill-current" : ""}`} />}
                                                            loading={likingIds.includes(item.id)}
                                                            onClick={() => void likeImage(item)}
                                                        >
                                                            {item.likeCount || 0}
                                                        </Button>
                                                        <Button size="small" icon={<MessageCircle className="size-3.5" />} onClick={() => openComments(item)}>
                                                            {item.commentCount || 0}
                                                        </Button>
                                                    </div>
                                                    {promptText ? (
                                                        <div className="rounded-md bg-secondary p-3 text-xs leading-5 text-secondary-foreground">
                                                            <p className={`${promptExpanded ? "whitespace-pre-wrap" : "line-clamp-3"} break-words`}>{promptText}</p>
                                                            {shouldFoldPrompt ? (
                                                                <button type="button" className="mt-2 text-xs font-medium text-primary hover:underline" onClick={() => togglePrompt(item.id)}>
                                                                    {promptExpanded ? "收起提示词" : "展开提示词"}
                                                                </button>
                                                            ) : null}
                                                        </div>
                                                    ) : null}
                                                </div>
                                            </article>
                                        );
                                    })}
                                </div>
                            </MaybeGalleryPreview>
                        ) : (
                            <div className="grid min-h-[520px] place-items-center rounded-lg border border-dashed border-border bg-card">
                                <Empty description={loading ? "正在读取画廊" : "暂无公开作品"} />
                            </div>
                        )}

                        {total > pageSize ? (
                            <div className="mt-8 flex justify-center">
                                <Pagination current={page} pageSize={pageSize} total={total} showSizeChanger={false} onChange={setPage} />
                            </div>
                        ) : null}
                    </div>
                </div>
            </section>
            <Modal
                title={<span className="block truncate pr-8">{commentImage ? `评论：${commentImage.title}` : "评论"}</span>}
                open={Boolean(commentImage)}
                width={640}
                onCancel={closeComments}
                onOk={() => void submitComment()}
                okText="发布评论"
                cancelText="关闭"
                confirmLoading={commentSubmitting}
                okButtonProps={{ disabled: !token || !commentText.trim() }}
                destroyOnHidden
            >
                <div className="space-y-4">
                    <Input.TextArea className="thin-scrollbar" value={commentText} rows={3} maxLength={500} showCount placeholder={token ? "写下你的评论" : "登录后可以评论"} disabled={!token || commentSubmitting} onChange={(event) => setCommentText(event.target.value)} />
                    <div className="flex items-center justify-between gap-3">
                        <span className="text-sm font-medium text-foreground">全部评论</span>
                        <span className="shrink-0 text-xs text-muted-foreground">{commentTotal} 条</span>
                    </div>
                    <div className="thin-scrollbar max-h-[min(52vh,420px)] overflow-y-auto rounded-lg border border-border bg-card">
                        {comments.length ? (
                            <div className="divide-y divide-border">
                                {comments.map((comment) => (
                                    <div key={comment.id} className="flex gap-3 p-3">
                                        <Avatar className="shrink-0" src={comment.avatarUrl} size={32}>
                                            {galleryAvatarText(comment)}
                                        </Avatar>
                                        <div className="min-w-0 flex-1">
                                            <div className="flex min-w-0 items-baseline justify-between gap-3">
                                                <span className="truncate text-sm font-medium text-foreground">{galleryUserName(comment)}</span>
                                                <time className="shrink-0 text-xs text-muted-foreground" dateTime={comment.createdAt}>
                                                    {formatGalleryDate(comment.createdAt)}
                                                </time>
                                            </div>
                                            <p className="mt-1 whitespace-pre-wrap break-words text-sm leading-6 text-foreground">{comment.content}</p>
                                        </div>
                                    </div>
                                ))}
                                {hasMoreComments ? (
                                    <div className="p-3">
                                        <Button block size="small" loading={commentsLoading} onClick={() => commentImage && void loadComments(commentImage, commentPage + 1)}>
                                            加载更多评论
                                        </Button>
                                    </div>
                                ) : null}
                            </div>
                        ) : (
                            <div className="grid min-h-40 place-items-center px-4 py-8">
                                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={commentsLoading ? "正在读取评论" : "暂无评论"} />
                            </div>
                        )}
                    </div>
                </div>
            </Modal>
        </main>
    );
}

function formatGalleryDate(value: string) {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value || "-";
    return new Intl.DateTimeFormat("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" }).format(date);
}

function GallerySortLabel({ icon, text }: { icon: ReactNode; text: string }) {
    return (
        <span className="inline-flex items-center justify-center gap-1.5">
            {icon}
            {text}
        </span>
    );
}

function mergeGalleryImage(item: GalleryImage, patch: Partial<GalleryImage>) {
    return { ...item, ...patch, imageUrl: item.imageUrl.startsWith("blob:") ? item.imageUrl : patch.imageUrl || item.imageUrl };
}

function compareGalleryLikes(a: GalleryImage, b: GalleryImage) {
    if (a.recommended !== b.recommended) return Number(b.recommended) - Number(a.recommended);
    if ((a.likeCount || 0) !== (b.likeCount || 0)) return (b.likeCount || 0) - (a.likeCount || 0);
    return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
}

function MaybeGalleryPreview({ enabled, children }: { enabled: boolean; children: ReactNode }) {
    return enabled ? <Image.PreviewGroup>{children}</Image.PreviewGroup> : <>{children}</>;
}

function galleryUserName(user: Pick<GalleryImage | GalleryComment, "username" | "displayName">) {
    return user.username || user.displayName || "用户";
}

function galleryAvatarText(user: Pick<GalleryImage | GalleryComment, "username" | "displayName">) {
    return galleryUserName(user).slice(0, 1);
}
