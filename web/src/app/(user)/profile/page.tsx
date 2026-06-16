"use client";

import { CalendarDays, Heart, ImageIcon, MessageCircle, RefreshCw, UserRound } from "lucide-react";
import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import { App, Avatar, Button, Empty, Image, Pagination, Segmented, Tag } from "antd";

import {
    fetchMyGalleryRewardStats,
    fetchMyLikedGalleryImages,
    fetchMyReceivedLikeGalleryImages,
    toggleGalleryLike,
    type GalleryImage,
    type GalleryQuery,
    type GalleryRewardStats,
} from "@/services/api/gallery";
import { useUserStore } from "@/stores/use-user-store";

const pageSize = 12;
type ProfileGalleryTab = "liked" | "received";
type ProfileGallerySort = NonNullable<GalleryQuery["sort"]>;

const tabOptions: { label: ReactNode; value: ProfileGalleryTab }[] = [
    { value: "liked", label: <ProfileOption icon={<Heart className="size-3.5" />} text="我点赞过" /> },
    { value: "received", label: <ProfileOption icon={<ImageIcon className="size-3.5" />} text="收到点赞" /> },
];

const sortOptions: { label: ReactNode; value: ProfileGallerySort }[] = [
    { value: "time", label: <ProfileOption icon={<CalendarDays className="size-3.5" />} text="最近" /> },
    { value: "likes", label: <ProfileOption icon={<Heart className="size-3.5" />} text="点赞" /> },
];

export default function ProfilePage() {
    const { message } = App.useApp();
    const token = useUserStore((state) => state.token);
    const user = useUserStore((state) => state.user);
    const isReady = useUserStore((state) => state.isReady);
    const hydrateUser = useUserStore((state) => state.hydrateUser);
    const [activeTab, setActiveTab] = useState<ProfileGalleryTab>("liked");
    const [sort, setSort] = useState<ProfileGallerySort>("time");
    const [items, setItems] = useState<GalleryImage[]>([]);
    const [total, setTotal] = useState(0);
    const [page, setPage] = useState(1);
    const [loading, setLoading] = useState(false);
    const [likingIds, setLikingIds] = useState<string[]>([]);
    const [refreshNonce, setRefreshNonce] = useState(0);
    const [rewardStatsNonce, setRewardStatsNonce] = useState(0);
    const [rewardStats, setRewardStats] = useState<GalleryRewardStats | null>(null);
    const userName = user?.displayName || user?.username || "用户";
    const avatarText = userName.slice(0, 1).toUpperCase();
    const joinedAt = useMemo(() => formatProfileDate(user?.createdAt || ""), [user?.createdAt]);
    const emptyText = activeTab === "liked" ? "还没有点赞过公开作品" : "你的公开作品还没有收到点赞";

    useEffect(() => {
        if (token) void hydrateUser();
    }, [hydrateUser, token]);

    useEffect(() => {
        if (!token || !user) return;
        let active = true;
        void fetchMyGalleryRewardStats(token)
            .then((data) => {
                if (active) setRewardStats(data);
            })
            .catch((error) => message.error(error instanceof Error ? error.message : "读取今日奖励统计失败"));
        return () => {
            active = false;
        };
    }, [message, rewardStatsNonce, token, user?.id]);

    useEffect(() => {
        if (!token || !user) return;
        let active = true;
        const query = { page, pageSize, sort };
        setLoading(true);
        void (activeTab === "liked" ? fetchMyLikedGalleryImages(token, query) : fetchMyReceivedLikeGalleryImages(token, query))
            .then((data) => {
                if (!active) return;
                setItems(data.items);
                setTotal(data.total);
            })
            .catch((error) => message.error(error instanceof Error ? error.message : "读取个人主页失败"))
            .finally(() => {
                if (active) setLoading(false);
            });
        return () => {
            active = false;
        };
    }, [activeTab, message, page, refreshNonce, sort, token, user]);

    const updateImage = (image: GalleryImage, liked: boolean) => {
        setItems((value) => {
            if ((activeTab === "liked" && !liked) || (activeTab === "received" && (image.likeCount || 0) <= 0)) {
                return value.filter((item) => item.id !== image.id);
            }
            const next = value.map((item) => (item.id === image.id ? mergeGalleryImage(item, image) : item));
            if (sort === "likes") next.sort(compareGalleryLikes);
            return next;
        });
        if ((activeTab === "liked" && !liked) || (activeTab === "received" && (image.likeCount || 0) <= 0)) {
            setTotal((value) => Math.max(0, value - 1));
        }
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
            updateImage(result.image, result.liked);
            void hydrateUser();
            setRewardStatsNonce((value) => value + 1);
            showGalleryRewardMessage(message, result.rewardCredits || result.image.rewardCredits);
        } catch (error) {
            message.error(error instanceof Error ? error.message : "点赞失败");
        } finally {
            setLikingIds((value) => value.filter((id) => id !== item.id));
        }
    };

    if (!isReady) {
        return (
            <main className="grid h-full place-items-center overflow-y-auto bg-background px-4 text-foreground">
                <Empty description="正在读取账号资料" />
            </main>
        );
    }

    if (isReady && !user) {
        return (
            <main className="grid h-full place-items-center overflow-y-auto bg-background px-4 text-foreground">
                <div className="w-full max-w-md rounded-lg border border-border bg-card p-6 text-center">
                    <div className="mx-auto grid size-12 place-items-center rounded-md bg-secondary text-primary">
                        <UserRound className="size-6" />
                    </div>
                    <h1 className="mt-4 text-2xl font-semibold">登录后查看个人主页</h1>
                    <p className="mt-2 text-sm leading-6 text-muted-foreground">这里会汇总你的账号资料、点赞过的作品和公开作品收到的点赞。</p>
                    <Button className="mt-5" type="primary" href="/login?redirect=%2Fprofile">
                        登录账号
                    </Button>
                </div>
            </main>
        );
    }

    return (
        <main className="h-full overflow-y-auto bg-background text-foreground">
            <section className="mx-auto max-w-[1600px] px-4 py-6 sm:px-6 lg:py-8">
                <div className="rounded-lg border border-border bg-card p-5 sm:p-6">
                    <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
                        <div className="flex min-w-0 items-center gap-4">
                            <Avatar size={72} src={user?.avatarUrl || undefined} className="shrink-0 border border-border bg-background text-xl font-semibold text-foreground">
                                {avatarText}
                            </Avatar>
                            <div className="min-w-0">
                                <div className="flex min-w-0 flex-wrap items-center gap-2">
                                    <h1 className="truncate text-2xl font-semibold text-foreground sm:text-3xl">{userName}</h1>
                                    <Tag className="m-0">{user?.role === "admin" ? "管理员" : "创作者"}</Tag>
                                </div>
                                <div className="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-muted-foreground">
                                    <span>@{user?.username || "-"}</span>
                                    <span>加入于 {joinedAt}</span>
                                </div>
                            </div>
                        </div>
                        <div className="grid gap-4 sm:grid-cols-3 lg:min-w-[420px]">
                            <ProfileStat label="算力点" value={(user?.credits || 0).toLocaleString()} />
                            <ProfileStat label="我点赞过" value={(user?.likeGivenCount || 0).toLocaleString()} />
                            <ProfileStat label="收到点赞" value={(user?.likeReceivedCount || 0).toLocaleString()} />
                        </div>
                    </div>
                </div>

                <div className="mt-5 rounded-lg border border-border bg-card p-5 sm:p-6">
                    <div className="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
                        <div>
                            <h2 className="text-lg font-semibold text-foreground">今日奖励</h2>
                            <p className="mt-1 text-sm text-muted-foreground">{rewardStats?.date || "-"} 的画廊互动奖励统计</p>
                        </div>
                        <Button size="small" icon={<RefreshCw className="size-3.5" />} onClick={() => setRewardStatsNonce((value) => value + 1)}>
                            刷新
                        </Button>
                    </div>
                    <div className="mt-4 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                        <ProfileStat label="上传奖励次数" value={(rewardStats?.uploadRewardCount || 0).toLocaleString()} />
                        <ProfileStat label="点赞奖励次数" value={(rewardStats?.likeRewardCount || 0).toLocaleString()} />
                        <ProfileStat label="被点赞奖励次数" value={(rewardStats?.receivedLikeRewardCount || 0).toLocaleString()} />
                        <ProfileStat label="被点赞获得点数" value={(rewardStats?.receivedLikeRewardCredits || 0).toLocaleString()} />
                    </div>
                </div>

                <div className="mt-5 rounded-lg border border-border bg-card">
                    <div className="flex flex-col gap-3 border-b border-border px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
                        <Segmented<ProfileGalleryTab>
                            value={activeTab}
                            options={tabOptions}
                            onChange={(value) => {
                                setActiveTab(value);
                                setPage(1);
                                setSort(value === "liked" ? "time" : "likes");
                            }}
                        />
                        <div className="flex flex-wrap items-center gap-2">
                            <Segmented<ProfileGallerySort>
                                value={sort}
                                options={sortOptions}
                                onChange={(value) => {
                                    setSort(value);
                                    setPage(1);
                                }}
                            />
                            <Button icon={<RefreshCw className="size-4" />} loading={loading} onClick={() => setRefreshNonce((value) => value + 1)}>
                                刷新
                            </Button>
                        </div>
                    </div>

                    <div className="p-4 sm:p-5">
                        {items.length ? (
                            <Image.PreviewGroup>
                                <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
                                    {items.map((item) => (
                                        <ProfileGalleryCard key={item.id} item={item} loading={likingIds.includes(item.id)} onLike={() => void likeImage(item)} />
                                    ))}
                                </div>
                            </Image.PreviewGroup>
                        ) : (
                            <div className="grid min-h-[360px] place-items-center rounded-lg border border-dashed border-border bg-background">
                                <Empty description={loading ? "正在读取作品" : emptyText} />
                            </div>
                        )}

                        {total > pageSize ? (
                            <div className="mt-6 flex justify-center">
                                <Pagination current={page} pageSize={pageSize} total={total} showSizeChanger={false} onChange={setPage} />
                            </div>
                        ) : null}
                    </div>
                </div>
            </section>
        </main>
    );
}

function ProfileGalleryCard({ item, loading, onLike }: { item: GalleryImage; loading: boolean; onLike: () => void }) {
    return (
        <article className="overflow-hidden rounded-lg border border-border bg-card">
            <div className="relative aspect-[4/3] overflow-hidden bg-muted">
                <Image src={item.imageUrl} alt={item.title} width="100%" height="100%" style={{ display: "block", height: "100%", objectFit: "cover" }} preview={{ mask: "放大" }} />
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
                    <span className="min-w-0 flex-1 truncate">{galleryUserName(item)}</span>
                    <span className="shrink-0">{formatProfileDate(item.createdAt)}</span>
                </div>
                <div className="flex items-center gap-2 border-t border-border pt-3">
                    <Button size="small" type={item.liked ? "primary" : "default"} icon={<Heart className={`size-3.5 ${item.liked ? "fill-current" : ""}`} />} loading={loading} onClick={onLike}>
                        {item.likeCount || 0}
                    </Button>
                    <span className="inline-flex h-7 items-center gap-1.5 rounded-md border border-border px-2 text-xs text-muted-foreground">
                        <MessageCircle className="size-3.5" />
                        {item.commentCount || 0}
                    </span>
                    <span className="ml-auto max-w-[45%] truncate text-xs text-muted-foreground">{item.model}</span>
                </div>
            </div>
        </article>
    );
}

function ProfileStat({ label, value }: { label: string; value: string }) {
    return (
        <div className="min-w-0 border-t border-border pt-3 sm:border-l sm:border-t-0 sm:pl-4 sm:pt-0">
            <div className="text-sm text-muted-foreground">{label}</div>
            <div className="mt-1 truncate text-2xl font-semibold tabular-nums text-foreground">{value}</div>
        </div>
    );
}

function ProfileOption({ icon, text }: { icon: ReactNode; text: string }) {
    return (
        <span className="inline-flex items-center justify-center gap-1.5">
            {icon}
            {text}
        </span>
    );
}

function formatProfileDate(value: string) {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return "-";
    return new Intl.DateTimeFormat("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" }).format(date);
}

function mergeGalleryImage(item: GalleryImage, patch: GalleryImage) {
    return { ...item, ...patch, imageUrl: item.imageUrl.startsWith("blob:") ? item.imageUrl : patch.imageUrl || item.imageUrl };
}

function compareGalleryLikes(a: GalleryImage, b: GalleryImage) {
    if ((a.likeCount || 0) !== (b.likeCount || 0)) return (b.likeCount || 0) - (a.likeCount || 0);
    return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
}

function galleryUserName(user: Pick<GalleryImage, "username" | "displayName">) {
    return user.username || user.displayName || "用户";
}

function galleryAvatarText(user: Pick<GalleryImage, "username" | "displayName">) {
    return galleryUserName(user).slice(0, 1);
}

function showGalleryRewardMessage(message: ReturnType<typeof App.useApp>["message"], credits?: number) {
    if ((credits || 0) > 0) message.success(`获得 ${credits} 算力点奖励`);
}
