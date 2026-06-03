"use client";

import { DeleteOutlined, EditOutlined, EyeOutlined, ReloadOutlined, SearchOutlined } from "@ant-design/icons";
import { ProTable, type ProColumns } from "@ant-design/pro-components";
import { Button, Card, Col, Flex, Form, Image, Input, Modal, Row, Select, Space, Switch, Tag, Tooltip, Typography } from "antd";
import { useEffect, useState } from "react";

import type { GalleryImage } from "@/services/api/gallery";
import { useAdminGallery } from "./use-admin-gallery";

type GalleryFormValues = Partial<GalleryImage> & { tagText?: string };

const statusOptions = [
    { label: "全部状态", value: "" },
    { label: "公开", value: "public" },
    { label: "下架", value: "hidden" },
    { label: "删除", value: "deleted" },
];

const statusLabels: Record<GalleryImage["status"], { label: string; color: string }> = {
    public: { label: "公开", color: "green" },
    hidden: { label: "下架", color: "orange" },
    deleted: { label: "删除", color: "red" },
};

export default function AdminGalleryPage() {
    const { images, tags, keyword, status, tag, page, pageSize, total, isLoading, searchImages, changeStatus, changeTag, changePage, changePageSize, resetFilters, refreshImages, saveImage, setStatus } = useAdminGallery();
    const [form] = Form.useForm<GalleryFormValues>();
    const [keywordText, setKeywordText] = useState(keyword);
    const [editingImage, setEditingImage] = useState<GalleryImage | null>(null);
    const [detailImage, setDetailImage] = useState<GalleryImage | null>(null);
    const [statusImage, setStatusImage] = useState<{ item: GalleryImage; status: GalleryImage["status"] } | null>(null);
    const tagOptions = tags.map((item) => ({ label: item, value: item }));

    useEffect(() => setKeywordText(keyword), [keyword]);
    useEffect(() => {
        if (editingImage) form.setFieldsValue({ ...editingImage, tagText: editingImage.tags?.join(", ") || "" });
    }, [editingImage, form]);

    const save = async () => {
        if (!editingImage) return;
        const value = await form.validateFields();
        await saveImage({
            ...editingImage,
            ...value,
            tags: (value.tagText || "")
                .split(",")
                .map((item) => item.trim())
                .filter(Boolean),
        });
        setEditingImage(null);
    };

    const columns: ProColumns<GalleryImage>[] = [
        {
            title: "图片",
            dataIndex: "imageUrl",
            width: 92,
            render: (_, item) => <Image src={item.imageUrl} alt={item.title} width={58} height={58} style={{ objectFit: "cover", borderRadius: 8 }} preview={{ mask: "放大" }} />,
        },
        {
            title: "作品",
            dataIndex: "title",
            width: 260,
            render: (_, item) => (
                <Flex vertical style={{ minWidth: 0 }}>
                    <Typography.Link strong ellipsis onClick={() => setDetailImage(item)}>
                        {item.title}
                    </Typography.Link>
                    <Typography.Text type="secondary" ellipsis>
                        {item.model}
                    </Typography.Text>
                </Flex>
            ),
        },
        {
            title: "状态",
            dataIndex: "status",
            width: 92,
            render: (_, item) => <Tag color={statusLabels[item.status]?.color}>{statusLabels[item.status]?.label || item.status}</Tag>,
        },
        {
            title: "提示词",
            dataIndex: "showPrompt",
            width: 92,
            render: (_, item) => <Tag>{item.showPrompt ? "公开" : "隐藏"}</Tag>,
        },
        {
            title: "标签",
            dataIndex: "tags",
            width: 180,
            render: (_, item) => (
                <Space size={[4, 4]} wrap>
                    {(item.tags || []).slice(0, 3).map((tag) => (
                        <Tag key={tag}>{tag}</Tag>
                    ))}
                </Space>
            ),
        },
        {
            title: "推荐",
            dataIndex: "recommended",
            width: 80,
            render: (_, item) => <Tag color={item.recommended ? "gold" : "default"}>{item.recommended ? "是" : "否"}</Tag>,
        },
        {
            title: "操作",
            key: "actions",
            width: 156,
            align: "right",
            render: (_, item) => (
                <Space size={4}>
                    <Tooltip title="详情">
                        <Button type="text" size="small" icon={<EyeOutlined />} onClick={() => setDetailImage(item)} />
                    </Tooltip>
                    <Tooltip title="编辑">
                        <Button type="text" size="small" icon={<EditOutlined />} onClick={() => setEditingImage(item)} />
                    </Tooltip>
                    {item.status === "public" ? (
                        <Button size="small" onClick={() => setStatusImage({ item, status: "hidden" })}>
                            下架
                        </Button>
                    ) : item.status === "hidden" ? (
                        <Button size="small" onClick={() => setStatusImage({ item, status: "public" })}>
                            恢复
                        </Button>
                    ) : null}
                    <Tooltip title="删除">
                        <Button danger type="text" size="small" icon={<DeleteOutlined />} disabled={item.status === "deleted"} onClick={() => setStatusImage({ item, status: "deleted" })} />
                    </Tooltip>
                </Space>
            ),
        },
    ];

    return (
        <main className="admin-page-shell">
            <Flex vertical gap={16}>
                <Card className="admin-filter-card" variant="borderless">
                    <Form layout="vertical">
                        <Row gutter={16} align="bottom">
                            <Col flex="360px">
                                <Form.Item label="关键词">
                                    <Input.Search value={keywordText} placeholder="搜索标题、描述、提示词或模型" allowClear enterButton={<SearchOutlined />} onSearch={() => searchImages(keywordText)} onChange={(event) => setKeywordText(event.target.value)} />
                                </Form.Item>
                            </Col>
                            <Col flex="180px">
                                <Form.Item label="状态">
                                    <Select value={status} onChange={changeStatus} options={statusOptions} />
                                </Form.Item>
                            </Col>
                            <Col flex="220px">
                                <Form.Item label="标签">
                                    <Select mode="multiple" allowClear maxTagCount="responsive" value={tag} onChange={changeTag} options={tagOptions} placeholder="全部标签" />
                                </Form.Item>
                            </Col>
                            <Col flex="none">
                                <Form.Item>
                                    <Space>
                                        <Button
                                            onClick={() => {
                                                setKeywordText("");
                                                resetFilters();
                                            }}
                                        >
                                            重置
                                        </Button>
                                        <Button type="primary" icon={<ReloadOutlined />} onClick={() => searchImages(keywordText)}>
                                            查询
                                        </Button>
                                    </Space>
                                </Form.Item>
                            </Col>
                        </Row>
                    </Form>
                </Card>

                <ProTable<GalleryImage>
                    rowKey="id"
                    columns={columns}
                    dataSource={images}
                    loading={isLoading}
                    search={false}
                    defaultSize="middle"
                    tableLayout="fixed"
                    cardProps={{ className: "admin-table-card", variant: "borderless" }}
                    headerTitle={
                        <Space>
                            <Typography.Text strong>画廊作品</Typography.Text>
                            <Tag>{total} 张</Tag>
                        </Space>
                    }
                    options={{ density: true, setting: true, reload: () => void refreshImages() }}
                    pagination={{
                        current: page,
                        pageSize,
                        total,
                        showSizeChanger: true,
                        pageSizeOptions: [10, 20, 50, 100],
                        showTotal: (value) => `共 ${value} 张`,
                        onChange: (nextPage, nextPageSize) => (nextPageSize !== pageSize ? changePageSize(nextPageSize) : changePage(nextPage)),
                    }}
                />
            </Flex>

            <Modal title="编辑画廊作品" open={Boolean(editingImage)} width={760} onCancel={() => setEditingImage(null)} onOk={() => void save()} okText="保存" cancelText="取消" destroyOnHidden>
                <Form form={form} layout="vertical" requiredMark={false}>
                    <Form.Item name="title" label="标题" rules={[{ required: true, message: "请输入标题" }]}>
                        <Input />
                    </Form.Item>
                    <Form.Item name="description" label="描述">
                        <Input.TextArea rows={3} />
                    </Form.Item>
                    <Form.Item name="tagText" label="标签，用逗号分隔">
                        <Input />
                    </Form.Item>
                    <Form.Item name="status" label="状态">
                        <Select options={statusOptions.slice(1)} />
                    </Form.Item>
                    <Form.Item name="showPrompt" label="公开提示词" valuePropName="checked">
                        <Switch />
                    </Form.Item>
                    <Form.Item name="recommended" label="推荐" valuePropName="checked">
                        <Switch />
                    </Form.Item>
                </Form>
            </Modal>

            <Modal title="画廊作品详情" open={Boolean(detailImage)} width={820} onCancel={() => setDetailImage(null)} footer={<Button onClick={() => setDetailImage(null)}>关闭</Button>}>
                {detailImage ? (
                    <Flex vertical gap={14}>
                        <Image src={detailImage.imageUrl} alt={detailImage.title} width="100%" style={{ maxHeight: 420, objectFit: "contain", borderRadius: 8 }} />
                        <Typography.Title level={5} style={{ margin: 0 }}>
                            {detailImage.title}
                        </Typography.Title>
                        <Space wrap>
                            <Tag color={statusLabels[detailImage.status]?.color}>{statusLabels[detailImage.status]?.label}</Tag>
                            <Tag>{detailImage.showPrompt ? "公开提示词" : "隐藏提示词"}</Tag>
                            <Tag>{detailImage.model}</Tag>
                            {(detailImage.tags || []).map((tag) => (
                                <Tag key={tag}>{tag}</Tag>
                            ))}
                        </Space>
                        {detailImage.description ? <Typography.Paragraph>{detailImage.description}</Typography.Paragraph> : null}
                        <Input.TextArea value={detailImage.prompt} rows={5} readOnly />
                    </Flex>
                ) : null}
            </Modal>

            <Modal
                title="更新作品状态"
                open={Boolean(statusImage)}
                onCancel={() => setStatusImage(null)}
                onOk={async () => {
                    if (!statusImage) return;
                    await setStatus(statusImage.item.id, statusImage.status);
                    setStatusImage(null);
                }}
                okText="确认"
                okButtonProps={{ danger: statusImage?.status === "deleted" }}
                cancelText="取消"
            >
                确认将「{statusImage?.item.title}」状态改为「{statusImage ? statusLabels[statusImage.status].label : ""}」吗？
            </Modal>
        </main>
    );
}
