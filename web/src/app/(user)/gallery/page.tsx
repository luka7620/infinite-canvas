"use client";

import { Search, SlidersHorizontal } from "lucide-react";
import { useEffect, useState } from "react";
import { App, Empty, Image, Input, Pagination, Tag } from "antd";

import { fetchGalleryImages, type GalleryImage } from "@/services/api/gallery";

const pageSize = 24;

export default function GalleryPage() {
    const { message } = App.useApp();
    const [items, setItems] = useState<GalleryImage[]>([]);
    const [tags, setTags] = useState<string[]>([]);
    const [total, setTotal] = useState(0);
    const [page, setPage] = useState(1);
    const [keyword, setKeyword] = useState("");
    const [keywordText, setKeywordText] = useState("");
    const [selectedTags, setSelectedTags] = useState<string[]>([]);
    const [expandedPromptIds, setExpandedPromptIds] = useState<string[]>([]);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        setLoading(true);
        void fetchGalleryImages({ keyword, tag: selectedTags, page, pageSize })
            .then((data) => {
                setItems(data.items);
                setTags(data.tags);
                setTotal(data.total);
            })
            .catch((error) => message.error(error instanceof Error ? error.message : "读取画廊失败"))
            .finally(() => setLoading(false));
    }, [keyword, message, page, selectedTags]);

    const toggleTag = (tag: string) => {
        setPage(1);
        setSelectedTags((value) => (value.includes(tag) ? value.filter((item) => item !== tag) : [...value, tag]));
    };

    const togglePrompt = (id: string) => {
        setExpandedPromptIds((value) => (value.includes(id) ? value.filter((item) => item !== id) : [...value, id]));
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
                            <Image.PreviewGroup>
                                <div className="columns-1 gap-4 sm:columns-2 xl:columns-3 2xl:columns-4">
                                    {items.map((item) => {
                                        const promptText = item.showPrompt && item.prompt ? item.prompt : "";
                                        const promptExpanded = expandedPromptIds.includes(item.id);
                                        const shouldFoldPrompt = promptText.length > 120;

                                        return (
                                            <article key={item.id} className="mb-4 break-inside-avoid overflow-hidden rounded-lg border border-border bg-card">
                                                <Image src={item.imageUrl} alt={item.title} width="100%" style={{ display: "block" }} preview={{ mask: "放大" }} />
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
                                                    <div className="text-xs text-muted-foreground">{item.model}</div>
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
                            </Image.PreviewGroup>
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
        </main>
    );
}
