"use client";

import { BookOpen, ClipboardPaste, Download, FolderPlus, History, ImagePlus, LoaderCircle, PenLine, Plus, SlidersHorizontal, Sparkles, Trash2, Upload } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { App, Button, Drawer, Empty, Image, Input, Modal, Switch, Tag, Typography } from "antd";
import { saveAs } from "file-saver";

import { ImageSettingsPanel } from "@/components/image-settings-panel";
import { ModelPicker } from "@/components/model-picker";
import { PromptSelectDialog } from "@/components/prompts/prompt-select-dialog";
import { AssetPickerModal, type InsertAssetPayload } from "@/app/(user)/canvas/components/asset-picker-modal";
import {
    LogPanel,
    type GenerationLog,
    type GenerationLogConfig,
    saveGenerationLog,
    deleteGenerationLogs,
    readStoredLogs,
} from "@/components/generation-log-panel";
import { canvasThemes } from "@/lib/canvas-theme";
import { useConfigStore, useEffectiveConfig, type AiConfig } from "@/stores/use-config-store";
import { useThemeStore } from "@/stores/use-theme-store";
import { nanoid } from "nanoid";
import { formatBytes, formatDuration, getDataUrlByteSize, readImageMeta } from "@/lib/image-utils";
import { requestEdit, requestGeneration } from "@/services/api/image";
import { fetchGeneratedImageBlob, publishGalleryImage } from "@/services/api/gallery";
import { clearCachedGalleryLists } from "@/services/gallery-cache";
import { deleteStoredImages, resolveImageUrl, uploadImage, type UploadedImage } from "@/services/image-storage";
import { useAssetStore } from "@/stores/use-asset-store";
import { useUserStore } from "@/stores/use-user-store";
import type { ReferenceImage } from "@/types/image";

type GeneratedImage = {
    id: string;
    dataUrl: string;
    storageKey?: string;
    durationMs: number;
    width: number;
    height: number;
    bytes: number;
    mimeType?: string;
    generatedImageId: string;
};

type GenerationResult = {
    id: string;
    status: "pending" | "success" | "failed";
    image?: GeneratedImage;
    error?: string;
};

type UpdateAiConfig = <K extends keyof AiConfig>(key: K, value: AiConfig[K]) => void;

const IMAGE_GENERATION_CONCURRENCY = 3;

export default function ImagePage() {
    const { message } = App.useApp();
    const fileInputRef = useRef<HTMLInputElement>(null);
    const config = useConfigStore((state) => state.config);
    const effectiveConfig = useEffectiveConfig();
    const updateConfig = useConfigStore((state) => state.updateConfig);
    const isAiConfigReady = useConfigStore((state) => state.isAiConfigReady);
    const openConfigDialog = useConfigStore((state) => state.openConfigDialog);
    const addAsset = useAssetStore((state) => state.addAsset);
    const token = useUserStore((state) => state.token);
    const hydrateUser = useUserStore((state) => state.hydrateUser);
    const [prompt, setPrompt] = useState("");
    const [references, setReferences] = useState<ReferenceImage[]>([]);
    const [results, setResults] = useState<GenerationResult[]>([]);
    const [logs, setLogs] = useState<GenerationLog[]>([]);
    const [running, setRunning] = useState(false);
    const [logsOpen, setLogsOpen] = useState(false);
    const [settingsOpen, setSettingsOpen] = useState(false);
    const [promptDialogOpen, setPromptDialogOpen] = useState(false);
    const [assetPickerOpen, setAssetPickerOpen] = useState(false);
    const [startedAt, setStartedAt] = useState(0);
    const [elapsedMs, setElapsedMs] = useState(0);
    const [selectedLogIds, setSelectedLogIds] = useState<string[]>([]);
    const [previewLog, setPreviewLog] = useState<GenerationLog | null>(null);
    const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
    const [publishingImage, setPublishingImage] = useState<GeneratedImage | null>(null);
    const [publishTitle, setPublishTitle] = useState("");
    const [publishDescription, setPublishDescription] = useState("");
    const [publishTags, setPublishTags] = useState("");
    const [publishShowPrompt, setPublishShowPrompt] = useState(false);
    const [publishLoading, setPublishLoading] = useState(false);
    const abortControllerRef = useRef<AbortController | null>(null);

    const model = effectiveConfig.imageModel || effectiveConfig.model;
    const canGenerate = Boolean(prompt.trim());
    const generationCount = Math.max(1, Math.min(10, Number(config.count) || 1));

    useEffect(() => {
        if (!running || !startedAt) return;
        const timer = window.setInterval(() => setElapsedMs(performance.now() - startedAt), 1000);
        return () => window.clearInterval(timer);
    }, [running, startedAt]);

    useEffect(() => {
        void refreshLogs();
    }, []);

    useEffect(() => {
        if (!running) return;
        const handler = (e: BeforeUnloadEvent) => {
            e.preventDefault();
            e.returnValue = "";
        };
        window.addEventListener("beforeunload", handler);
        return () => window.removeEventListener("beforeunload", handler);
    }, [running]);

    const addReferences = async (files?: FileList | null) => {
        const imageFiles = Array.from(files || []).filter((file) => file.type.startsWith("image/"));
        const nextReferences = await Promise.all(
            imageFiles.map(async (file) => {
                const image = await uploadImage(file);
                return { id: nanoid(), name: file.name, type: image.mimeType, dataUrl: image.url, storageKey: image.storageKey };
            }),
        );
        setReferences((value) => [...value, ...nextReferences]);
    };

    const addReferencesFromClipboard = async () => {
        try {
            const items = await navigator.clipboard.read();
            const blobs = await Promise.all(items.flatMap((item) => item.types.filter((type) => type.startsWith("image/")).map((type) => item.getType(type))));
            if (!blobs.length) {
                message.error("剪切板里没有可读取的图片");
                return;
            }
            const nextReferences = await Promise.all(
                blobs.map(async (blob, index) => {
                    const image = await uploadImage(blob);
                    return { id: nanoid(), name: `clipboard-${index + 1}.png`, type: image.mimeType, dataUrl: image.url, storageKey: image.storageKey };
                }),
            );
            setReferences((value) => [...value, ...nextReferences]);
            message.success(`已读取 ${nextReferences.length} 张参考图`);
        } catch {
            message.error("剪切板里没有可读取的图片");
        }
    };

    const generate = async () => {
        const text = prompt.trim();
        if (!text) {
            message.error("请输入生图提示词");
            return;
        }
        if (!isAiConfigReady(effectiveConfig, model)) {
            message.warning("请先完成配置");
            openConfigDialog(true);
            return;
        }

        const snapshot = buildRequestSnapshot();
        if (!snapshot) return;

        const ac = new AbortController();
        abortControllerRef.current = ac;
        const signal = ac.signal;

        setElapsedMs(0);
        setRunning(true);
        setPreviewLog(null);
        setResults(Array.from({ length: generationCount }, () => ({ id: nanoid(), status: "pending" })));
        const batchStartedAt = performance.now();
        setStartedAt(batchStartedAt);
        const logId = nanoid();
        const logCreatedAt = Date.now();
        const logTime = new Date().toLocaleString("zh-CN", { hour12: false });
        const logImages: Array<GeneratedImage | undefined> = [];
        const failedIndexes = new Set<number>();
        let logSaveQueue = Promise.resolve();
        const getLogImages = () => logImages.filter((image): image is GeneratedImage => Boolean(image));
        const persistLog = () => {
            const images = getLogImages();
            const log = buildLog({
                id: logId,
                createdAt: logCreatedAt,
                time: logTime,
                prompt: text,
                model,
                config: { ...snapshot.config, count: String(generationCount) },
                references: snapshot.references,
                durationMs: performance.now() - batchStartedAt,
                successCount: images.length,
                failCount: failedIndexes.size,
                status: images.length ? "成功" : "失败",
                images,
            });
            logSaveQueue = logSaveQueue
                .catch(() => undefined)
                .then(async () => {
                    await saveGenerationLog(log);
                    await refreshLogs();
                });
            return logSaveQueue.catch(() => undefined);
        };
        const cacheGeneratedImage = async (index: number, image: GeneratedImage) => {
            try {
                const stored = await storeGeneratedImage(image);
                const storedImage = applyStoredImage(image, stored);
                logImages[index] = storedImage;
                setResults((value) => updateResultAt(value, index, { image: storedImage }));
            } catch {
                logImages[index] = image;
            }
            failedIndexes.delete(index);
            await persistLog();
        };
        const cacheGenerationFailure = async (index: number) => {
            failedIndexes.add(index);
            await persistLog();
        };
        await persistLog();

        const result = await runWithConcurrency(generationCount, IMAGE_GENERATION_CONCURRENCY, (index) => runGenerationSlot(index, snapshot, signal, cacheGeneratedImage, cacheGenerationFailure));
        const successCount = result.filter((item): item is PromiseFulfilledResult<GeneratedImage> => item.status === "fulfilled").length;
        const failCount = generationCount - successCount;
        const failed = result.find((item): item is PromiseRejectedResult => item.status === "rejected");

        try {
            if (failCount) Array.from({ length: generationCount }).forEach((_, index) => (result[index]?.status === "rejected" ? failedIndexes.add(index) : undefined));
            await persistLog();
            successCount ? message.success("图片已生成") : message.error(failed?.reason instanceof Error ? failed.reason.message : "生成失败");
        } catch {
            // log save failed — ignore
        } finally {
            if (abortControllerRef.current === ac) abortControllerRef.current = null;
            setRunning(false);
        }
    };

    const downloadImage = (image: GeneratedImage, index: number) => {
        saveAs(image.dataUrl, `image-${index + 1}.png`);
    };

    const storeGeneratedImage = async (image: GeneratedImage): Promise<UploadedImage> => {
        if (image.storageKey) {
            const url = await resolveImageUrl(image.storageKey, image.dataUrl);
            if (url) {
                return {
                    url,
                    storageKey: image.storageKey,
                    width: image.width,
                    height: image.height,
                    bytes: image.bytes,
                    mimeType: image.mimeType || "image/png",
                };
            }
        }
        try {
            return await uploadImage(image.dataUrl);
        } catch (error) {
            if (!token || !image.generatedImageId) throw error;
            return uploadImage(await fetchGeneratedImageBlob(token, image.generatedImageId));
        }
    };

    const addResultToReferences = async (image: GeneratedImage, index: number) => {
        try {
            const stored = await storeGeneratedImage(image);
            const storedImage = applyStoredImage(image, stored);
            setResults((value) => updateResultAt(value, index, { image: storedImage }));
            setReferences((value) => [...value, { id: nanoid(), name: `result-${index + 1}.png`, type: stored.mimeType, dataUrl: stored.url, storageKey: stored.storageKey }]);
            message.success("已加入参考图");
        } catch (error) {
            message.error(error instanceof Error ? error.message : "加入参考图失败");
        }
    };

    const saveResultToAssets = async (image: GeneratedImage, index: number) => {
        try {
            const stored = await storeGeneratedImage(image);
            const storedImage = applyStoredImage(image, stored);
            setResults((value) => updateResultAt(value, index, { image: storedImage }));
            addAsset({
                kind: "image",
                title: `生成结果 ${index + 1}`,
                coverUrl: stored.url,
                tags: [],
                source: "生图工作台",
                data: { dataUrl: stored.url, storageKey: stored.storageKey, width: stored.width, height: stored.height, bytes: stored.bytes, mimeType: stored.mimeType },
                metadata: { source: "image-page", prompt },
            });
            message.success("已加入我的素材");
        } catch (error) {
            message.error(error instanceof Error ? error.message : "加入素材失败");
        }
    };

    const openPublishDialog = (image: GeneratedImage, index: number) => {
        if (!image.generatedImageId) {
            message.warning("这张图片没有生成记录，暂不能上传到画廊");
            return;
        }
        setPublishingImage(image);
        setPublishTitle(`生成作品 ${index + 1}`);
        setPublishDescription("");
        setPublishTags("");
        setPublishShowPrompt(false);
    };

    const publishImageToGallery = async () => {
        if (!token) {
            message.warning("请先登录");
            return;
        }
        if (!publishingImage?.generatedImageId) {
            message.warning("这张图片没有生成记录，暂不能上传到画廊");
            return;
        }
        setPublishLoading(true);
        try {
            const result = await publishGalleryImage(token, {
                generatedImageId: publishingImage.generatedImageId,
                title: publishTitle,
                description: publishDescription,
                tags: publishTags
                    .split(",")
                    .map((tag) => tag.trim())
                    .filter(Boolean),
                showPrompt: publishShowPrompt,
            });
            await clearCachedGalleryLists().catch(() => undefined);
            void hydrateUser();
            showPublishRewardMessage(message, result.rewardCredits);
            setPublishingImage(null);
        } catch (error) {
            message.error(error instanceof Error ? error.message : "上传失败");
        } finally {
            setPublishLoading(false);
        }
    };

    const insertPickedAsset = async (payload: InsertAssetPayload) => {
        if (payload.kind === "text") {
            setPrompt(payload.content);
        } else {
            const stored = await uploadImage(payload.dataUrl);
            setReferences((value) => [...value, { id: nanoid(), name: payload.title, type: stored.mimeType, dataUrl: stored.url, storageKey: stored.storageKey }]);
        }
        setAssetPickerOpen(false);
    };

    const doCreateSession = () => {
        if (abortControllerRef.current) {
            abortControllerRef.current.abort();
            abortControllerRef.current = null;
        }
        setRunning(false);
        setPrompt("");
        setReferences([]);
        setResults([]);
        setElapsedMs(0);
        setStartedAt(0);
        setSelectedLogIds([]);
        setPreviewLog(null);
    };

    const createSession = () => {
        if (running || abortControllerRef.current) {
            message.warning("当前有图片正在生成，请等待结束后再新建。");
            return;
        }
        doCreateSession();
    };

    const deleteSelectedLogs = () => {
        const imageKeys = logs.filter((log) => selectedLogIds.includes(log.id)).flatMap((log) => log.images.map((image) => image.storageKey).filter((key): key is string => Boolean(key)));
        void Promise.all([deleteStoredImages(imageKeys), deleteGenerationLogs(selectedLogIds)]).then(refreshLogs);
        if (previewLog && selectedLogIds.includes(previewLog.id)) {
            setPreviewLog(null);
            setResults([]);
        }
        setSelectedLogIds([]);
        setDeleteConfirmOpen(false);
    };

    const refreshLogs = async () => setLogs(await readStoredLogs());

    const previewGenerationLog = async (log: GenerationLog) => {
        setPreviewLog(log);
        setLogsOpen(false);
        setPrompt(log.prompt);
        setReferences(log.references || []);
        if (log.config.imageModel || log.model) updateConfig("imageModel", log.config.imageModel || log.model);
        if (log.config.quality) updateConfig("quality", log.config.quality);
        if (log.config.size) updateConfig("size", log.config.size);
        if (log.config.count) updateConfig("count", log.config.count);
        setResults(log.images.map((image) => ({ id: image.id, status: "success" as const, image: { ...image, generatedImageId: image.generatedImageId || "" } })));
    };

    const buildRequestSnapshot = () => {
        const text = prompt.trim();
        if (!text) {
            message.error("请输入生图提示词");
            return null;
        }
        if (!isAiConfigReady(effectiveConfig, model)) {
            message.warning("请先完成配置");
            openConfigDialog(true);
            return null;
        }
        return { text, config: { ...effectiveConfig, model, count: "1" }, references: [...references] };
    };

    const runGenerationSlot = async (
        index: number,
        snapshot: { text: string; config: AiConfig; references: ReferenceImage[] },
        signal?: AbortSignal,
        onSuccess?: (index: number, image: GeneratedImage) => Promise<void>,
        onFailure?: (index: number, error: unknown) => Promise<void>,
    ) => {
        const itemStartedAt = performance.now();
        try {
            const result = snapshot.references.length ? await requestEdit(snapshot.config, snapshot.text, snapshot.references, "image-page", signal) : await requestGeneration(snapshot.config, snapshot.text, "image-page", signal);
            const image = result[0];
            if (!image) throw new Error("接口没有返回图片");
            const meta = await readImageMeta(image.dataUrl);
            const nextImage = { id: image.id, dataUrl: image.dataUrl, generatedImageId: image.generatedImageId, durationMs: performance.now() - itemStartedAt, width: meta.width, height: meta.height, bytes: image.dataUrl.startsWith("data:") ? getDataUrlByteSize(image.dataUrl) : 0 };
            setResults((value) => updateResultAt(value, index, { status: "success", image: nextImage }));
            if (onSuccess) await onSuccess(index, nextImage).catch(() => undefined);
            return nextImage;
        } catch (error) {
            setResults((value) => updateResultAt(value, index, { status: "failed", error: error instanceof Error ? error.message : "生成失败" }));
            if (onFailure) await onFailure(index, error).catch(() => undefined);
            throw error;
        }
    };

    const retryResult = (index: number) => {
        const snapshot = buildRequestSnapshot();
        if (!snapshot) return;
        setPreviewLog(null);
        setResults((value) => updateResultAt(value, index, { status: "pending", error: undefined, image: undefined }));
        void runGenerationSlot(index, snapshot).catch(() => {});
    };

    return (
        <div className="flex h-full flex-col overflow-hidden bg-background text-foreground">
            <main className="grid min-h-0 flex-1 grid-cols-1 gap-4 overflow-y-auto p-4 lg:grid-cols-[300px_minmax(0,1fr)] lg:overflow-hidden xl:grid-cols-[320px_minmax(0,1fr)]">
                <aside className="thin-scrollbar hidden min-h-0 overflow-y-auto rounded-lg border border-border bg-card p-4 lg:block">
                    <LogPanel
                        logs={logs}
                        selectedLogIds={selectedLogIds}
                        activeLogId={previewLog?.id}
                        running={running}
                        onSelectedLogIdsChange={setSelectedLogIds}
                        onCreateSession={createSession}
                        onDeleteSelected={() => setDeleteConfirmOpen(true)}
                        onPreviewLog={(log) => void previewGenerationLog(log)}
                    />
                </aside>

                <section className="grid gap-3 lg:min-h-0 lg:overflow-hidden xl:grid-cols-[420px_minmax(0,1fr)]">
                    <div className="thin-scrollbar flex flex-col rounded-lg border border-border bg-card p-4 lg:min-h-0 lg:overflow-y-auto">
                        <div>
                            <div className="flex items-start justify-between gap-3">
                                <div className="min-w-0">
                                    <h1 className="text-2xl font-semibold text-foreground">生图工作台</h1>
                                    <p className="mt-1 text-sm leading-6 text-muted-foreground">独立生成、保存素材，并发布到作品画廊。</p>
                                </div>
                                <div className="flex shrink-0 gap-2 lg:hidden">
                                    <Button icon={<History className="size-4" />} onClick={() => setLogsOpen(true)}>
                                        记录
                                    </Button>
                                    <Button icon={<SlidersHorizontal className="size-4" />} onClick={() => setSettingsOpen(true)}>
                                        参数
                                    </Button>
                                </div>
                            </div>
                        </div>

                        <div className="mt-6 space-y-5">
                            <div>
                                <div className="mb-2 flex items-center justify-between gap-3">
                                    <span className="text-base font-semibold">提示词</span>
                                    <div className="flex gap-2">
                                        <Button size="small" icon={<BookOpen className="size-3.5" />} onClick={() => setPromptDialogOpen(true)}>
                                            查看提示词库
                                        </Button>
                                        <Button size="small" icon={<FolderPlus className="size-3.5" />} onClick={() => setAssetPickerOpen(true)}>
                                            查看我的素材
                                        </Button>
                                    </div>
                                </div>
                                <Input.TextArea value={prompt} onChange={(event) => setPrompt(event.target.value)} rows={7} placeholder="描述画面主体、风格、构图、光线和用途" />
                            </div>

                            <div className="min-w-0">
                                <div className="mb-2 flex items-center justify-between gap-3">
                                    <span className="text-base font-semibold">参考图</span>
                                    <div className="flex gap-2">
                                        <Button size="small" icon={<ClipboardPaste className="size-3.5" />} onClick={() => void addReferencesFromClipboard()}>
                                            剪切板
                                        </Button>
                                        <Button size="small" icon={<Upload className="size-3.5" />} onClick={() => fileInputRef.current?.click()}>
                                            上传
                                        </Button>
                                    </div>
                                </div>
                                <div
                                    className="hover-scrollbar hover-scrollbar-hint flex min-h-24 w-full min-w-0 max-w-full gap-2 overflow-x-scroll overflow-y-hidden rounded-lg border border-dashed border-border p-2 pb-3 overscroll-x-contain"
                                    onWheel={(event) => {
                                        if (event.currentTarget.scrollWidth <= event.currentTarget.clientWidth) return;
                                        event.preventDefault();
                                        event.currentTarget.scrollLeft += event.deltaY;
                                    }}
                                >
                                    {references.map((item) => (
                                        <div key={item.id} className="group relative size-20 shrink-0 overflow-hidden rounded-md border border-border">
                                            <img src={item.dataUrl} alt={item.name} className="size-full object-cover" />
                                            <button
                                                type="button"
                                                className="absolute right-1 top-1 hidden size-6 items-center justify-center rounded bg-black/60 text-white group-hover:flex"
                                                onClick={() => setReferences((value) => value.filter((ref) => ref.id !== item.id))}
                                                aria-label="移除参考图"
                                            >
                                                <Trash2 className="size-3.5" />
                                            </button>
                                        </div>
                                    ))}
                                    {!references.length ? <div className="flex min-w-full items-center justify-center text-sm text-muted-foreground">暂无参考图</div> : null}
                                </div>
                            </div>

                            <div className="flex items-center justify-between rounded-lg border border-border bg-secondary px-3 py-2 text-sm sm:hidden">
                                <span className="truncate text-muted-foreground">
                                    {model} · {effectiveConfig.size} · {effectiveConfig.quality}
                                </span>
                                <Button size="small" type="text" icon={<SlidersHorizontal className="size-4" />} onClick={() => setSettingsOpen(true)}>
                                    调整
                                </Button>
                            </div>

                            <div className="hidden gap-4 sm:grid sm:grid-cols-2">
                                <GenerationSettings config={effectiveConfig} model={model} updateConfig={updateConfig} openConfigDialog={openConfigDialog} />
                            </div>
                        </div>

                        <div className="mt-auto pt-6">
                            <Button type="primary" size="large" block icon={<Sparkles className="size-4" />} loading={running} disabled={!canGenerate || running} onClick={() => void generate()}>
                                开始生成
                            </Button>
                        </div>
                    </div>

                    <div className="thin-scrollbar rounded-lg border border-border bg-card p-4 lg:min-h-0 lg:overflow-y-auto lg:p-5">
                        <div className="mb-4 flex items-center justify-between gap-3">
                            <div>
                                <h2 className="text-xl font-semibold">生成结果</h2>
                            </div>
                            {running ? <Tag className="m-0 px-2 py-1">等待 {formatDuration(elapsedMs)}</Tag> : null}
                        </div>
                        {results.length ? (
                            <div className="grid gap-4 sm:grid-cols-2 2xl:grid-cols-3">
                                {results.map((result, index) =>
                                    result.status === "success" && result.image ? (
                                        <ResultImageCard key={result.id} image={result.image} index={index} onEdit={addResultToReferences} onDownload={downloadImage} onSaveAsset={saveResultToAssets} onPublish={openPublishDialog} />
                                    ) : result.status === "failed" ? (
                                        <FailedImageCard key={result.id} error={result.error || "生成失败"} onRetry={() => retryResult(index)} />
                                    ) : (
                                        <PendingImageCard key={result.id} />
                                    ),
                                )}
                            </div>
                        ) : (
                            <div className="flex min-h-[320px] flex-col items-center justify-center rounded-lg border border-dashed border-border text-center lg:min-h-[560px]">
                                <ImagePlus className="mb-4 size-11 text-muted-foreground" />
                                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有生成图片" />
                            </div>
                        )}
                    </div>
                </section>
            </main>
            <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                multiple
                className="hidden"
                onChange={(event) => {
                    void addReferences(event.target.files);
                    event.target.value = "";
                }}
            />
            <Drawer title="生成记录" placement="bottom" size="large" open={logsOpen} onClose={() => setLogsOpen(false)}>
                <LogPanel
                    logs={logs}
                    selectedLogIds={selectedLogIds}
                    activeLogId={previewLog?.id}
                    running={running}
                    onSelectedLogIdsChange={setSelectedLogIds}
                    onCreateSession={createSession}
                    onDeleteSelected={() => setDeleteConfirmOpen(true)}
                    onPreviewLog={(log) => void previewGenerationLog(log)}
                />
            </Drawer>
            <Drawer title="参数" placement="bottom" size="82vh" open={settingsOpen} onClose={() => setSettingsOpen(false)}>
                <div className="grid grid-cols-2 gap-3 pb-4">
                    <GenerationSettings config={effectiveConfig} model={model} updateConfig={updateConfig} openConfigDialog={openConfigDialog} />
                </div>
            </Drawer>
            <PromptSelectDialog open={promptDialogOpen} onOpenChange={setPromptDialogOpen} onSelect={setPrompt} />
            <AssetPickerModal open={assetPickerOpen} defaultTab="my-assets" onInsert={(payload) => void insertPickedAsset(payload)} onClose={() => setAssetPickerOpen(false)} />
            <Modal title="删除生成记录" open={deleteConfirmOpen} onCancel={() => setDeleteConfirmOpen(false)} onOk={deleteSelectedLogs} okText="删除" okButtonProps={{ danger: true }} cancelText="取消">
                确定删除选中的 {selectedLogIds.length} 条生成记录吗？
            </Modal>
            <Modal title="上传到画廊" open={Boolean(publishingImage)} confirmLoading={publishLoading} onCancel={() => setPublishingImage(null)} onOk={() => void publishImageToGallery()} okText="上传" cancelText="取消">
                <div className="space-y-4">
                    <p className="text-sm text-muted-foreground">仅支持上传使用云端链接的 GPT 模型生成图片。</p>
                    <label className="block">
                        <span className="mb-1.5 block text-sm font-medium">标题</span>
                        <Input value={publishTitle} maxLength={60} onChange={(event) => setPublishTitle(event.target.value)} />
                    </label>
                    <label className="block">
                        <span className="mb-1.5 block text-sm font-medium">描述</span>
                        <Input.TextArea value={publishDescription} rows={3} onChange={(event) => setPublishDescription(event.target.value)} />
                    </label>
                    <label className="block">
                        <span className="mb-1.5 block text-sm font-medium">标签，用逗号分隔</span>
                        <Input value={publishTags} onChange={(event) => setPublishTags(event.target.value)} />
                    </label>
                    <div className="flex items-center justify-between rounded-lg border border-border px-3 py-2">
                        <div>
                            <div className="text-sm font-medium">分享提示词</div>
                            <div className="text-xs text-muted-foreground">关闭后画廊前台不会展示本次生成提示词。</div>
                        </div>
                        <Switch checked={publishShowPrompt} onChange={setPublishShowPrompt} />
                    </div>
                </div>
            </Modal>
        </div>
    );
}

function GenerationSettings({ config, model, updateConfig, openConfigDialog }: { config: AiConfig; model: string; updateConfig: UpdateAiConfig; openConfigDialog: (shouldPromptContinue?: boolean) => void }) {
    const theme = canvasThemes[useThemeStore((state) => state.theme)];

    return (
        <>
            <label className="col-span-2 block min-w-0 sm:col-span-1">
                <span className="mb-1.5 block text-sm font-semibold sm:mb-2 sm:text-base">模型</span>
                <ModelPicker config={config} value={model} onChange={(value) => updateConfig("imageModel", value)} fullWidth onMissingConfig={() => openConfigDialog(false)} />
            </label>
            <div className="col-span-2">
                <ImageSettingsPanel config={config} onConfigChange={(key, value) => updateConfig(key, value)} theme={theme} showTitle={false} className="space-y-4" maxCount={10} />
            </div>
        </>
    );
}

function ResultImageCard({
    image,
    index,
    onEdit,
    onDownload,
    onSaveAsset,
    onPublish,
}: {
    image: GeneratedImage;
    index: number;
    onEdit: (image: GeneratedImage, index: number) => void;
    onDownload: (image: GeneratedImage, index: number) => void;
    onSaveAsset: (image: GeneratedImage, index: number) => void;
    onPublish: (image: GeneratedImage, index: number) => void;
}) {
    return (
        <div className="overflow-hidden rounded-lg border border-border bg-card">
            <Image src={image.dataUrl} alt={`生成结果 ${index + 1}`} className="aspect-square object-cover" />
            <div className="flex flex-wrap items-center justify-between gap-x-3 gap-y-2 border-t border-border px-3 py-2.5">
                <div className="flex min-w-0 flex-wrap gap-x-2 gap-y-1 text-xs text-muted-foreground">
                    <span>
                        {image.width}x{image.height}
                    </span>
                    <span>{formatBytes(image.bytes)}</span>
                    <span>{formatDuration(image.durationMs)}</span>
                </div>
                <div className="flex shrink-0 gap-1">
                    <Button size="small" icon={<FolderPlus className="size-3.5" />} onClick={() => void onSaveAsset(image, index)}>
                        添加到素材
                    </Button>
                    <Button size="small" icon={<Sparkles className="size-3.5" />} disabled={!image.generatedImageId} onClick={() => onPublish(image, index)}>
                        上传画廊
                    </Button>
                    <Button size="small" icon={<PenLine className="size-3.5" />} onClick={() => void onEdit(image, index)}>
                        加入参考图
                    </Button>
                    <Button size="small" icon={<Download className="size-3.5" />} onClick={() => onDownload(image, index)}>
                        下载
                    </Button>
                </div>
            </div>
        </div>
    );
}

function PendingImageCard() {
    return (
        <div className="relative aspect-square overflow-hidden rounded-lg border border-dashed border-border bg-secondary">
            <div
                className="absolute inset-0 opacity-60"
                style={{
                    backgroundImage: "radial-gradient(circle, color-mix(in oklch, currentColor 26%, transparent) 1.4px, transparent 1.6px)",
                    backgroundSize: "16px 16px",
                }}
            />
            <div className="absolute inset-0 flex flex-col items-center justify-center gap-2 text-sm text-muted-foreground">
                <LoaderCircle className="size-6 animate-spin" />
                <span>生成中</span>
            </div>
        </div>
    );
}

function FailedImageCard({ error, onRetry }: { error: string; onRetry: () => void }) {
    return (
        <div className="overflow-hidden rounded-lg border border-red-200 bg-red-50 dark:border-red-950 dark:bg-red-950/20">
            <div className="flex aspect-square flex-col items-center justify-center gap-3 p-5 text-center">
                <div className="text-sm font-medium text-red-600 dark:text-red-300">生成失败</div>
                <Typography.Paragraph ellipsis={{ rows: 4 }} className="!mb-0 !text-xs !text-red-500 dark:!text-red-300">
                    {error}
                </Typography.Paragraph>
            </div>
            <div className="flex justify-end border-t border-red-200 p-3 dark:border-red-950">
                <Button size="small" danger onClick={onRetry}>
                    重试
                </Button>
            </div>
        </div>
    );
}

function updateResultAt(results: GenerationResult[], index: number, next: Partial<GenerationResult>) {
    return results.map((item, itemIndex) => (itemIndex === index ? { ...item, ...next } : item));
}

function applyStoredImage(image: GeneratedImage, stored: UploadedImage): GeneratedImage {
    return { ...image, dataUrl: stored.url, storageKey: stored.storageKey, width: stored.width, height: stored.height, bytes: stored.bytes, mimeType: stored.mimeType };
}

async function runWithConcurrency<T>(count: number, limit: number, task: (index: number) => Promise<T>): Promise<PromiseSettledResult<T>[]> {
    const results: PromiseSettledResult<T>[] = [];
    let nextIndex = 0;
    const workerCount = Math.min(count, limit);
    await Promise.all(
        Array.from({ length: workerCount }, async () => {
            for (;;) {
                const index = nextIndex;
                nextIndex += 1;
                if (index >= count) return;
                try {
                    results[index] = { status: "fulfilled", value: await task(index) };
                } catch (reason) {
                    results[index] = { status: "rejected", reason };
                }
            }
        }),
    );
    return results;
}

function buildLog({
    id,
    createdAt,
    time,
    prompt,
    model,
    config,
    references,
    durationMs,
    successCount,
    failCount,
    status,
    images,
}: {
    id?: string;
    createdAt?: number;
    time?: string;
    prompt: string;
    model: string;
    config: GenerationLogConfig;
    references: ReferenceImage[];
    durationMs: number;
    successCount: number;
    failCount: number;
    status: GenerationLog["status"];
    images: GeneratedImage[];
}): GenerationLog {
    const logConfig = {
        model: config.model,
        imageModel: config.imageModel,
        quality: config.quality,
        size: config.size,
        count: config.count,
    };
    return {
        id: id || nanoid(),
        createdAt: createdAt || Date.now(),
        title: prompt.slice(0, 12) || "未命名",
        prompt,
        time: time || new Date().toLocaleString("zh-CN", { hour12: false }),
        model,
        config: logConfig,
        references,
        durationMs,
        successCount,
        failCount,
        imageCount: Number(logConfig.count) || successCount,
        size: logConfig.size,
        quality: logConfig.quality,
        status,
        images,
        thumbnails: images.map((image) => image.dataUrl),
    };
}

function showPublishRewardMessage(message: ReturnType<typeof App.useApp>["message"], credits?: number) {
    message.success((credits || 0) > 0 ? `已上传到画廊，获得 ${credits} 算力点奖励` : "已上传到画廊");
}
