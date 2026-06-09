"use client";

import { ArrowRight, ImagePlus, Layers3, Library, LockKeyhole, Workflow } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { App, Button, Empty, Image, Tag } from "antd";

import { navigationTools, visibleNavigationTools } from "@/constant/navigation-tools";
import { cn } from "@/lib/utils";
import { fetchGalleryImages, type GalleryImage } from "@/services/api/gallery";
import { fetchPrompts, type Prompt } from "@/services/api/prompts";
import { useConfigStore } from "@/stores/use-config-store";
import { useUserStore } from "@/stores/use-user-store";

const workflow = [
    { label: "画布编排", value: "节点、连线、撤销重做", icon: Workflow },
    { label: "模型生成", value: "文生图、图生图、助手", icon: ImagePlus },
    { label: "素材沉淀", value: "本地素材与公开画廊", icon: Library },
];

export default function IndexPage() {
    const { message } = App.useApp();
    const token = useUserStore((state) => state.token);
    const videoEnabled = useConfigStore((state) => state.publicSettings?.features.videoEnabled === true);
    const tools = visibleNavigationTools(videoEnabled);
    const primaryTool = tools[0] || navigationTools[0];
    const [galleryShowcase, setGalleryShowcase] = useState<GalleryImage[]>([]);
    const [galleryReady, setGalleryReady] = useState(false);
    const [promptShowcase, setPromptShowcase] = useState<Prompt[]>([]);
    const [galleryPreviewIndex, setGalleryPreviewIndex] = useState(0);
    const [galleryPreviewOpen, setGalleryPreviewOpen] = useState(false);
    const [previewIndex, setPreviewIndex] = useState(0);
    const [previewOpen, setPreviewOpen] = useState(false);
    const heroGallery = galleryShowcase.slice(0, 5);
    const canViewGalleryImages = Boolean(token);

    useEffect(() => {
        setGalleryReady(false);
        void fetchGalleryImages({ pageSize: 12 })
            .then((data) => setGalleryShowcase(data.items))
            .catch((error) => message.error(error instanceof Error ? error.message : "读取画廊失败"))
            .finally(() => setGalleryReady(true));
    }, [message]);

    useEffect(() => {
        void fetchPrompts({ pageSize: 12 })
            .then((data) => setPromptShowcase(data.items))
            .catch((error) => message.error(error instanceof Error ? error.message : "获取提示词失败"));
    }, [message]);

    return (
        <main className="h-full overflow-y-auto bg-background text-foreground">
            <section className="mx-auto grid max-w-[1600px] gap-5 px-4 py-6 sm:px-6 lg:grid-cols-[minmax(420px,0.85fr)_minmax(620px,1.15fr)] lg:items-stretch lg:py-8">
                <div className="flex h-full flex-col justify-between rounded-lg border border-border bg-card p-6 lg:p-8">
                    <div>
                        <div className="mb-8 inline-flex items-center gap-2 rounded-md border border-border bg-secondary px-3 py-1.5 text-sm font-medium text-secondary-foreground">
                            <Layers3 className="size-4" />
                            开源图片创作工作台
                        </div>
                        <h1 className="max-w-3xl text-balance text-5xl font-semibold leading-[1.02] tracking-normal text-foreground sm:text-6xl">LukaLeng生图公益</h1>
                        <p className="mt-6 max-w-2xl text-pretty text-base leading-8 text-muted-foreground sm:text-lg">
                            在一个桌面工作区里整理提示词、连接参考图、批量生成图片，并把稳定结果沉淀到素材和画廊。
                        </p>
                        <div className="mt-8 flex flex-wrap items-center gap-3">
                            <Button type="primary" size="large" href={`/${primaryTool.slug}`} icon={<ArrowRight className="size-4" />} iconPlacement="end">
                                进入工作台
                            </Button>
                            <Button size="large" href="/gallery">
                                浏览画廊
                            </Button>
                        </div>
                    </div>

                    <div className="mt-12 grid gap-3 sm:grid-cols-3 lg:grid-cols-1 xl:grid-cols-3">
                        {workflow.map((item) => {
                            const Icon = item.icon;
                            return (
                                <div key={item.label} className="rounded-md border border-border bg-background/60 p-4">
                                    <Icon className="mb-4 size-5 text-primary" />
                                    <div className="font-medium text-foreground">{item.label}</div>
                                    <div className="mt-1 text-sm leading-6 text-muted-foreground">{item.value}</div>
                                </div>
                            );
                        })}
                    </div>
                </div>

                <div className="grid gap-4">
                    <div className="overflow-hidden rounded-lg border border-border bg-card">
                        <div className="flex items-center justify-between border-b border-border px-4 py-3">
                            <div>
                                <h2 className="text-base font-semibold text-foreground">画廊精选</h2>
                                <p className="mt-1 text-sm text-muted-foreground">点击作品可预览，进入画廊查看公开提示词和标签。</p>
                            </div>
                            <Button href="/gallery" size="small" icon={<ArrowRight className="size-3.5" />} iconPlacement="end">
                                查看画廊
                            </Button>
                        </div>
                        {heroGallery.length ? (
                            <div className="grid h-[360px] grid-cols-6 grid-rows-6 gap-px bg-border xl:h-[420px]">
                                {heroGallery.map((item, index) => (
                                    <button
                                        key={item.id}
                                        type="button"
                                        onClick={() => {
                                            if (!canViewGalleryImages) {
                                                window.location.href = "/login?redirect=%2Fgallery";
                                                return;
                                            }
                                            setGalleryPreviewIndex(index);
                                            setGalleryPreviewOpen(true);
                                        }}
                                        className={cn(
                                            "group relative min-h-0 overflow-hidden bg-muted text-left",
                                            canViewGalleryImages ? "cursor-pointer" : "cursor-default",
                                            index === 0 && "col-span-3 row-span-6",
                                            index === 1 && "col-span-3 row-span-3",
                                            index > 1 && "col-span-1 row-span-3",
                                        )}
                                    >
                                        <img src={item.imageUrl} alt={item.title} className={cn("h-full w-full object-cover transition duration-300", canViewGalleryImages ? "group-hover:scale-[1.025]" : "scale-105 blur-xl")} />
                                        {!canViewGalleryImages ? (
                                            <div className="absolute inset-0 grid place-items-center bg-background/20 p-3">
                                                <span className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground shadow-sm">
                                                    <LockKeyhole className="size-3.5" />
                                                    登录后查看
                                                </span>
                                            </div>
                                        ) : null}
                                        <div className="absolute inset-x-0 bottom-0 bg-black/62 p-4 text-white">
                                            <div className="mb-2 flex flex-wrap gap-1.5">
                                                {item.tags.slice(0, index === 0 ? 3 : 2).map((tag) => (
                                                    <Tag key={tag} className="m-0 border-white/20 bg-white/15 text-[11px] text-white">
                                                        {tag}
                                                    </Tag>
                                                ))}
                                            </div>
                                            <h3 className="line-clamp-1 text-sm font-semibold">{item.title}</h3>
                                            <p className={cn("mt-1 text-xs leading-5 text-white/78", index === 0 ? "line-clamp-3" : "line-clamp-2")}>{item.description || (item.showPrompt ? item.prompt : item.model)}</p>
                                        </div>
                                    </button>
                                ))}
                            </div>
                        ) : (
                            <div className="grid h-[360px] place-items-center xl:h-[420px]">
                                <Empty description={galleryReady ? "暂无公开作品" : "画廊作品加载中"} />
                            </div>
                        )}
                    </div>

                    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                        {tools.map((tool) => {
                            const Icon = tool.icon;
                            return (
                                <Link key={tool.slug} href={`/${tool.slug}`} className="group flex min-h-20 items-center gap-3 rounded-lg border border-border bg-card p-4 transition hover:border-primary/40 hover:bg-secondary">
                                    <span className="grid size-10 shrink-0 place-items-center rounded-md bg-secondary text-primary transition group-hover:bg-primary group-hover:text-primary-foreground">
                                        <Icon className="size-5" />
                                    </span>
                                    <span className="min-w-0">
                                        <span className="block font-medium text-foreground">{tool.label}</span>
                                        <span className="mt-1 block text-sm text-muted-foreground">{tool.slug === "canvas" ? "管理和打开本地项目" : tool.slug === "image" ? "独立批量生成图片" : tool.slug === "gallery" ? "查看公开作品" : "进入对应工作区"}</span>
                                    </span>
                                </Link>
                            );
                        })}
                    </div>
                </div>
            </section>

            {promptShowcase.length ? (
                <section className="mx-auto max-w-[1600px] px-4 pb-8 sm:px-6">
                    <div className="grid gap-4 border-t border-border pt-6 lg:grid-cols-[260px_1fr]">
                        <div>
                            <h2 className="text-2xl font-semibold text-foreground">最近同步的提示词</h2>
                            <p className="mt-2 text-sm leading-6 text-muted-foreground">保留原有提示词预览和图片查看能力，减少首页装饰，让内容更容易被扫描。</p>
                        </div>
                        <div className="grid auto-rows-[168px] gap-3 sm:grid-cols-2 lg:grid-cols-4">
                            {promptShowcase.slice(5).map((item) => (
                                <button
                                    key={item.id}
                                    type="button"
                                    onClick={() => {
                                        setPreviewIndex(promptShowcase.findIndex((prompt) => prompt.id === item.id));
                                        setPreviewOpen(true);
                                    }}
                                    className="group relative cursor-pointer overflow-hidden rounded-lg border border-border bg-card text-left"
                                >
                                    <img src={item.coverUrl} alt={item.title} className="h-full w-full object-cover transition duration-300 group-hover:scale-[1.025]" />
                                    <div className="absolute inset-x-0 bottom-0 bg-black/64 p-3 text-white">
                                        <div className="mb-2 flex flex-wrap gap-1.5">
                                            {item.tags.slice(0, 2).map((tag) => (
                                                <Tag key={tag} className="m-0 border-white/20 bg-white/15 text-[11px] text-white">
                                                    {tag}
                                                </Tag>
                                            ))}
                                        </div>
                                        <h3 className="line-clamp-1 text-sm font-medium">{item.title}</h3>
                                    </div>
                                </button>
                            ))}
                        </div>
                    </div>
                </section>
            ) : null}

            <Image.PreviewGroup
                preview={{
                    open: previewOpen,
                    current: previewIndex,
                    onOpenChange: setPreviewOpen,
                    onChange: setPreviewIndex,
                }}
            >
                <div className="hidden">
                    {promptShowcase.map((item) => (
                        <Image key={item.id} src={item.coverUrl} alt={item.title} />
                    ))}
                </div>
            </Image.PreviewGroup>

            {canViewGalleryImages ? (
                <Image.PreviewGroup
                    preview={{
                        open: galleryPreviewOpen,
                        current: galleryPreviewIndex,
                        onOpenChange: setGalleryPreviewOpen,
                        onChange: setGalleryPreviewIndex,
                    }}
                >
                    <div className="hidden">
                        {galleryShowcase.map((item) => (
                            <Image key={item.id} src={item.imageUrl} alt={item.title} />
                        ))}
                    </div>
                </Image.PreviewGroup>
            ) : null}
        </main>
    );
}
