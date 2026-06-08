"use client";

import { BarChart3, Database, GalleryHorizontal, Image as ImageIcon, KeyRound, RefreshCw, Sparkles, UserPlus, Users, Wallet, type LucideIcon } from "lucide-react";
import { Alert, Button, Card, Col, Empty, Flex, Progress, Row, Segmented, Skeleton, Space, Statistic, Tag, Tooltip, Typography, theme } from "antd";
import dayjs from "dayjs";
import type { CSSProperties, ReactNode } from "react";

import type { AdminStats, AdminStatsDistributionItem, AdminStatsRange, AdminStatsTrendDay } from "@/services/api/admin";
import { useAdminStats } from "./use-admin-stats";

const rangeOptions: { label: string; value: AdminStatsRange }[] = [
    { label: "近 7 天", value: "7d" },
    { label: "近 30 天", value: "30d" },
    { label: "全部", value: "all" },
];

const summaryCards: { key: keyof AdminStats["summary"]; title: string; icon: LucideIcon; suffix?: string }[] = [
    { key: "userTotal", title: "用户总数", icon: Users, suffix: "人" },
    { key: "newUsers", title: "范围内新增用户", icon: UserPlus, suffix: "人" },
    { key: "creditBalanceTotal", title: "当前算力点总余额", icon: Wallet },
    { key: "creditConsumed", title: "范围内算力点消耗", icon: Sparkles },
    { key: "generatedImages", title: "范围内生成图片", icon: ImageIcon, suffix: "张" },
    { key: "publicGalleryImages", title: "公开画廊作品", icon: GalleryHorizontal, suffix: "个" },
    { key: "promptTotal", title: "提示词数量", icon: BarChart3, suffix: "条" },
    { key: "assetTotal", title: "素材数量", icon: Database, suffix: "个" },
    { key: "inviteCodeTotal", title: "邀请码数量", icon: KeyRound, suffix: "个" },
    { key: "inviteCodeEnabled", title: "启用邀请码", icon: KeyRound, suffix: "个" },
];

const trendSeries: { key: keyof Omit<AdminStatsTrendDay, "date">; label: string; color: string }[] = [
    { key: "newUsers", label: "新增用户", color: "#0f6f68" },
    { key: "creditConsumed", label: "算力点消费", color: "#d97706" },
    { key: "creditRefunded", label: "算力点返还", color: "#2563eb" },
    { key: "generatedImages", label: "生成图片", color: "#7c3aed" },
    { key: "galleryPublishes", label: "画廊发布", color: "#dc2626" },
];

const distributionGroups: { key: keyof AdminStats["distributions"]; title: string; empty: string }[] = [
    { key: "userStatus", title: "用户状态", empty: "暂无用户状态数据" },
    { key: "loginSource", title: "登录来源", empty: "暂无登录来源数据" },
    { key: "creditLogType", title: "算力点日志类型", empty: "暂无算力点日志数据" },
    { key: "galleryStatus", title: "画廊状态", empty: "暂无画廊状态数据" },
    { key: "modelUsage", title: "模型使用量", empty: "暂无模型使用数据" },
    { key: "assetType", title: "素材类型", empty: "暂无素材类型数据" },
];

export default function AdminStatsPage() {
    const { token: antToken } = theme.useToken();
    const { stats, range, isLoading, isRefreshing, isError, errorMessage, changeRange, refreshStats } = useAdminStats();

    return (
        <main className="admin-page-shell">
            <Flex vertical gap={16}>
                <Card className="admin-filter-card" variant="borderless">
                    <Flex align="center" justify="space-between" gap={16} wrap="wrap">
                        <Flex vertical gap={4}>
                            <Typography.Text strong>数据统计</Typography.Text>
                            <Typography.Text type="secondary">{stats ? `${stats.startDate} 至 ${stats.endDate}` : "选择时间范围查看后台运营概览"}</Typography.Text>
                        </Flex>
                        <Space wrap>
                            <Segmented value={range} options={rangeOptions} onChange={(value) => changeRange(value as AdminStatsRange)} />
                            <Button icon={<RefreshCw size={16} />} loading={isRefreshing} onClick={() => void refreshStats()}>
                                刷新
                            </Button>
                        </Space>
                    </Flex>
                </Card>

                {isError ? (
                    <Alert type="error" showIcon message="统计数据读取失败" description={errorMessage || "请稍后重试"} action={<Button onClick={() => void refreshStats()}>重试</Button>} />
                ) : null}

                {isLoading ? <StatsSkeleton /> : stats ? <StatsContent stats={stats} antToken={antToken} /> : <Empty description="暂无统计数据" />}
            </Flex>
        </main>
    );
}

function StatsContent({ stats, antToken }: { stats: AdminStats; antToken: { colorFillSecondary: string; colorPrimary: string; colorBorder: string } }) {
    const hasTrend = stats.trends.some((item) => trendSeries.some((series) => Number(item[series.key]) > 0));
    const trendRangeText = stats.trendStartDate === stats.startDate && stats.trendEndDate === stats.endDate ? `${stats.trends.length} 天` : `${stats.trendStartDate} 至 ${stats.trendEndDate}`;
    return (
        <>
            <Row gutter={[16, 16]}>
                {summaryCards.map((item) => {
                    const Icon = item.icon;
                    const value = stats.summary[item.key];
                    return (
                        <Col key={item.key} xs={24} sm={12} lg={8} xl={6}>
                            <Card variant="borderless" style={panelStyle(antToken)}>
                                <Flex align="center" justify="space-between" gap={16}>
                                    <Statistic title={item.title} value={value} suffix={item.suffix} />
                                    <span style={{ ...summaryIconStyle, background: antToken.colorFillSecondary, color: antToken.colorPrimary }}>
                                        <Icon size={20} />
                                    </span>
                                </Flex>
                            </Card>
                        </Col>
                    );
                })}
            </Row>

            <Card variant="borderless" style={panelStyle(antToken)}>
                <Flex align="center" justify="space-between" gap={12} wrap="wrap" style={{ marginBottom: 16 }}>
                    <Space>
                        <Typography.Text strong>趋势变化</Typography.Text>
                        <Tag>{trendRangeText}</Tag>
                    </Space>
                    <Space size={12} wrap>
                        {trendSeries.map((series) => (
                            <Space key={series.key} size={6}>
                                <span style={{ width: 8, height: 8, borderRadius: 8, background: series.color }} />
                                <Typography.Text type="secondary">{series.label}</Typography.Text>
                            </Space>
                        ))}
                    </Space>
                </Flex>
                {hasTrend ? <TrendBars trends={stats.trends} /> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前时间范围暂无趋势数据" />}
            </Card>

            <Row gutter={[16, 16]}>
                {distributionGroups.map((group) => (
                    <Col key={group.key} xs={24} lg={12} xl={8}>
                        <DistributionPanel title={group.title} empty={group.empty} items={stats.distributions[group.key]} antToken={antToken} />
                    </Col>
                ))}
            </Row>

            <Card variant="borderless" style={panelStyle(antToken)}>
                <Flex vertical gap={8}>
                    <Typography.Text strong>近期提示</Typography.Text>
                    <StatusHint condition={stats.summary.creditConsumed > stats.summary.creditRefunded} good="范围内算力点消费高于返还，扣费链路整体正常。" fallback="范围内算力点返还不低于消费，建议关注模型失败或重试情况。" />
                    <StatusHint condition={stats.summary.publicGalleryImages > 0} good="公开画廊已有作品，可继续观察发布和互动增长。" fallback="公开画廊暂无作品，需要确认生成记录发布路径是否符合预期。" />
                    <StatusHint condition={stats.summary.inviteCodeEnabled > 0} good={`当前有 ${stats.summary.inviteCodeEnabled} 个启用邀请码。`} fallback="当前没有启用的邀请码，关闭公开注册时新用户无法注册。" />
                </Flex>
            </Card>
        </>
    );
}

function TrendBars({ trends }: { trends: AdminStatsTrendDay[] }) {
    const maxValue = Math.max(1, ...trends.flatMap((item) => trendSeries.map((series) => Number(item[series.key]) || 0)));
    return (
        <div style={{ overflowX: "auto", paddingBottom: 4 }}>
            <div style={{ display: "grid", gridTemplateColumns: `repeat(${trends.length}, minmax(36px, 1fr))`, gap: 8, minWidth: Math.min(Math.max(trends.length * 44, 360), 1280) }}>
                {trends.map((item) => (
                    <TooltipDay key={item.date} item={item}>
                        <div style={{ display: "flex", minWidth: 0, flexDirection: "column", alignItems: "center", gap: 8 }}>
                            <div style={trendStackStyle}>
                                {trendSeries.map((series) => {
                                    const value = Number(item[series.key]) || 0;
                                    return <span key={series.key} style={{ display: "block", width: "100%", height: value > 0 ? `${Math.max(2, (value / maxValue) * 100)}%` : 0, borderRadius: 4, background: series.color }} />;
                                })}
                            </div>
                            <Typography.Text type="secondary" style={{ fontSize: 12, whiteSpace: "nowrap" }}>
                                {dayjs(item.date).format("MM-DD")}
                            </Typography.Text>
                        </div>
                    </TooltipDay>
                ))}
            </div>
        </div>
    );
}

function TooltipDay({ item, children }: { item: AdminStatsTrendDay; children: ReactNode }) {
    const { token } = theme.useToken();
    const title = (
        <div style={{ color: token.colorText, display: "flex", flexDirection: "column", gap: 6, minWidth: 168 }}>
            <span style={{ color: token.colorText, fontSize: 13, fontWeight: 600 }}>{item.date}</span>
            {trendSeries.map((series) => (
                <span key={series.key} style={{ alignItems: "center", color: token.colorTextSecondary, display: "flex", fontSize: 12, gap: 16, justifyContent: "space-between" }}>
                    <span style={{ alignItems: "center", display: "inline-flex", gap: 6 }}>
                        <span style={{ background: series.color, borderRadius: 8, flex: "0 0 8px", height: 8, width: 8 }} />
                        {series.label}
                    </span>
                    <span style={{ color: token.colorText, fontVariantNumeric: "tabular-nums" }}>{item[series.key]}</span>
                </span>
            ))}
        </div>
    );
    return (
        <Tooltip
            title={title}
            styles={{
                container: {
                    background: token.colorBgElevated,
                    border: `1px solid ${token.colorBorder}`,
                    boxShadow: token.boxShadowSecondary,
                    padding: "10px 12px",
                },
            }}
        >
            {children}
        </Tooltip>
    );
}

function DistributionPanel({ title, empty, items, antToken }: { title: string; empty: string; items: AdminStatsDistributionItem[]; antToken: { colorBorder: string } }) {
    const total = items.reduce((sum, item) => sum + item.value, 0);
    return (
        <Card variant="borderless" title={title} style={panelStyle(antToken)}>
            {total > 0 ? (
                <Flex vertical gap={12}>
                    {items.map((item) => {
                        const percent = Math.round((item.value / total) * 100);
                        return (
                            <Flex key={item.key} vertical gap={4}>
                                <Flex align="center" justify="space-between" gap={12}>
                                    <Typography.Text ellipsis>{item.label}</Typography.Text>
                                    <Typography.Text type="secondary">{item.value}</Typography.Text>
                                </Flex>
                                <Progress percent={percent} showInfo={false} size="small" />
                            </Flex>
                        );
                    })}
                </Flex>
            ) : (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={empty} />
            )}
        </Card>
    );
}

function StatusHint({ condition, good, fallback }: { condition: boolean; good: string; fallback: string }) {
    return (
        <Flex align="center" gap={8}>
            <Tag color={condition ? "green" : "orange"}>{condition ? "正常" : "关注"}</Tag>
            <Typography.Text type="secondary">{condition ? good : fallback}</Typography.Text>
        </Flex>
    );
}

function StatsSkeleton() {
    return (
        <Flex vertical gap={16}>
            <Row gutter={[16, 16]}>
                {summaryCards.map((_, index) => (
                    <Col key={index} xs={24} sm={12} lg={8} xl={6}>
                        <Card variant="borderless" style={{ border: "1px solid var(--border)", borderRadius: 8 }}>
                            <Skeleton active paragraph={false} />
                        </Card>
                    </Col>
                ))}
            </Row>
            <Card variant="borderless" style={{ border: "1px solid var(--border)", borderRadius: 8 }}>
                <Skeleton active />
            </Card>
        </Flex>
    );
}

const summaryIconStyle: CSSProperties = {
    alignItems: "center",
    borderRadius: 8,
    display: "inline-flex",
    flex: "0 0 40px",
    height: 40,
    justifyContent: "center",
    width: 40,
};

const trendStackStyle: CSSProperties = {
    alignItems: "end",
    display: "grid",
    gridTemplateColumns: "repeat(5, 1fr)",
    gap: 3,
    height: 132,
    width: 34,
};

function panelStyle(token: { colorBorder: string }): CSSProperties {
    return { border: `1px solid ${token.colorBorder}`, borderRadius: 8, boxShadow: "none" };
}
