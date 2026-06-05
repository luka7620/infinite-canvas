"use client";

import { CopyOutlined, DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined, SearchOutlined } from "@ant-design/icons";
import { ProTable, type ProColumns } from "@ant-design/pro-components";
import { Button, Card, Col, Form, Input, InputNumber, Modal, Row, Select, Space, Switch, Tag, Tooltip, Typography } from "antd";
import dayjs from "dayjs";
import { useEffect, useState } from "react";

import type { AdminInviteCode } from "@/services/api/admin";
import { useCopyText } from "@/hooks/use-copy-text";
import { useAdminInviteCodes } from "./use-admin-invite-codes";

type InviteCodeFormValues = Partial<AdminInviteCode>;
type BatchInviteCodeFormValues = Pick<AdminInviteCode, "type" | "credits" | "maxUses" | "enabled" | "remark"> & { count: number };

const typeOptions = [
    { label: "注册", value: "register" },
    { label: "兑换额度", value: "credits" },
];

const typeLabels: Record<AdminInviteCode["type"], string> = {
    register: "注册",
    credits: "兑换额度",
};

export default function AdminInviteCodesPage() {
    const { inviteCodes, keyword, type, page, pageSize, total, isLoading, searchInviteCodes, changeType, changePage, changePageSize, resetFilters, refreshInviteCodes, saveInviteCode, batchCreateInviteCodes, deleteInviteCode } = useAdminInviteCodes();
    const copyText = useCopyText();
    const [form] = Form.useForm<InviteCodeFormValues>();
    const [batchForm] = Form.useForm<BatchInviteCodeFormValues>();
    const currentType = Form.useWatch("type", form);
    const currentBatchType = Form.useWatch("type", batchForm);
    const [keywordText, setKeywordText] = useState(keyword);
    const [editingItem, setEditingItem] = useState<Partial<AdminInviteCode> | null>(null);
    const [batchOpen, setBatchOpen] = useState(false);
    const [batchResult, setBatchResult] = useState<AdminInviteCode[]>([]);
    const [selectedInviteCodes, setSelectedInviteCodes] = useState<AdminInviteCode[]>([]);
    const [deletingItem, setDeletingItem] = useState<AdminInviteCode | null>(null);

    useEffect(() => setKeywordText(keyword), [keyword]);

    useEffect(() => {
        if (editingItem) form.setFieldsValue({ type: "register", credits: 0, maxUses: 1, enabled: true, ...editingItem });
    }, [editingItem, form]);

    useEffect(() => {
        if (batchOpen) batchForm.setFieldsValue({ type: "register", count: 20, credits: 0, maxUses: 1, enabled: true, remark: "" });
    }, [batchForm, batchOpen]);

    const save = async () => {
        const value = await form.validateFields();
        await saveInviteCode({ ...editingItem, ...value, code: value.code?.trim() });
        setEditingItem(null);
    };

    const copyCodes = (items: AdminInviteCode[], successText = "邀请码已复制") => {
        const text = items.map((item) => item.code).filter(Boolean).join("\n");
        if (!text) return;
        copyText(text, successText);
    };

    const batchCreate = async () => {
        const value = await batchForm.validateFields();
        const result = await batchCreateInviteCodes({ ...value, remark: value.remark?.trim() });
        setBatchResult(result.items);
    };

    const columns: ProColumns<AdminInviteCode>[] = [
        {
            title: "邀请码",
            dataIndex: "code",
            width: 180,
            render: (_, item) => <Typography.Text copyable strong>{item.code}</Typography.Text>,
        },
        {
            title: "类型",
            dataIndex: "type",
            width: 120,
            render: (_, item) => <Tag color={item.type === "credits" ? "blue" : "green"}>{typeLabels[item.type] || item.type}</Tag>,
        },
        {
            title: "额度",
            dataIndex: "credits",
            width: 100,
            render: (_, item) => (item.type === "credits" ? item.credits : "-"),
        },
        {
            title: "使用",
            dataIndex: "usedCount",
            width: 120,
            render: (_, item) => `${item.usedCount}/${item.maxUses > 0 ? item.maxUses : "不限"}`,
        },
        {
            title: "状态",
            dataIndex: "enabled",
            width: 100,
            render: (_, item) => <Tag color={item.enabled ? "success" : "default"}>{item.enabled ? "启用" : "停用"}</Tag>,
        },
        {
            title: "备注",
            dataIndex: "remark",
            ellipsis: true,
            render: (_, item) => <Typography.Text type="secondary">{item.remark || "-"}</Typography.Text>,
        },
        {
            title: "创建时间",
            dataIndex: "createdAt",
            width: 180,
            render: (_, item) => <Typography.Text type="secondary">{item.createdAt ? dayjs(item.createdAt).format("YYYY-MM-DD HH:mm:ss") : "-"}</Typography.Text>,
        },
        {
            title: "操作",
            key: "actions",
            width: 96,
            align: "right",
            render: (_, item) => (
                <Space size={4}>
                    <Tooltip title="编辑">
                        <Button type="text" size="small" icon={<EditOutlined />} onClick={() => setEditingItem(item)} />
                    </Tooltip>
                    <Tooltip title="删除">
                        <Button danger type="text" size="small" icon={<DeleteOutlined />} onClick={() => setDeletingItem(item)} />
                    </Tooltip>
                </Space>
            ),
        },
    ];

    return (
        <main className="admin-page-shell">
            <Space direction="vertical" size={16} style={{ width: "100%" }}>
                <Card className="admin-filter-card" variant="borderless">
                    <Form layout="vertical">
                        <Row gutter={16} align="bottom">
                            <Col flex="320px">
                                <Form.Item label="关键词">
                                    <Input.Search value={keywordText} placeholder="搜索邀请码或备注" allowClear enterButton={<SearchOutlined />} onSearch={() => searchInviteCodes(keywordText)} onChange={(event) => setKeywordText(event.target.value)} />
                                </Form.Item>
                            </Col>
                            <Col flex="180px">
                                <Form.Item label="类型">
                                    <Select value={type} options={[{ label: "全部", value: "" }, ...typeOptions]} onChange={changeType} />
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
                                        <Button type="primary" icon={<ReloadOutlined />} onClick={() => searchInviteCodes(keywordText)}>
                                            查询
                                        </Button>
                                    </Space>
                                </Form.Item>
                            </Col>
                        </Row>
                    </Form>
                </Card>
                <ProTable<AdminInviteCode>
                    rowKey="id"
                    columns={columns}
                    dataSource={inviteCodes}
                    loading={isLoading}
                    search={false}
                    defaultSize="middle"
                    tableLayout="fixed"
                    cardProps={{ className: "admin-table-card", variant: "borderless" }}
                    headerTitle={
                        <Space>
                            <Typography.Text strong>邀请码</Typography.Text>
                            <Tag>{total} 个</Tag>
                        </Space>
                    }
                    options={{ density: true, setting: true, reload: () => void refreshInviteCodes() }}
                    rowSelection={{
                        selectedRowKeys: selectedInviteCodes.map((item) => item.id),
                        onChange: (_, rows) => setSelectedInviteCodes(rows),
                    }}
                    toolBarRender={() => [
                        <Button key="copy" icon={<CopyOutlined />} disabled={!selectedInviteCodes.length} onClick={() => copyCodes(selectedInviteCodes, `已复制 ${selectedInviteCodes.length} 个邀请码`)}>
                            复制选中
                        </Button>,
                        <Button
                            key="batch"
                            icon={<PlusOutlined />}
                            onClick={() => {
                                setBatchResult([]);
                                setBatchOpen(true);
                            }}
                        >
                            批量生成
                        </Button>,
                        <Button key="add" type="primary" icon={<PlusOutlined />} onClick={() => setEditingItem({ type: "register", credits: 0, maxUses: 1, enabled: true })}>
                            新增
                        </Button>,
                    ]}
                    pagination={{
                        current: page,
                        pageSize,
                        total,
                        showSizeChanger: true,
                        pageSizeOptions: [10, 20, 50, 100],
                        showTotal: (value) => `共 ${value} 个`,
                        onChange: (nextPage, nextPageSize) => (nextPageSize !== pageSize ? changePageSize(nextPageSize) : changePage(nextPage)),
                    }}
                />
            </Space>

            <Modal title={editingItem?.id ? "编辑邀请码" : "新增邀请码"} open={Boolean(editingItem)} width={680} onCancel={() => setEditingItem(null)} onOk={() => void save()} okText="保存" cancelText="取消" destroyOnHidden>
                <Form form={form} layout="vertical" requiredMark={false}>
                    <Row gutter={14}>
                        <Col span={12}>
                            <Form.Item name="code" label="邀请码">
                                <Input placeholder="留空自动生成" />
                            </Form.Item>
                        </Col>
                        <Col span={12}>
                            <Form.Item name="type" label="类型" rules={[{ required: true, message: "请选择类型" }]}>
                                <Select options={typeOptions} />
                            </Form.Item>
                        </Col>
                        <Col span={12}>
                            <Form.Item name="credits" label="兑换额度" rules={currentType === "credits" ? [{ required: true, message: "请输入兑换额度" }] : []}>
                                <InputNumber min={0} precision={0} disabled={currentType !== "credits"} style={{ width: "100%" }} />
                            </Form.Item>
                        </Col>
                        <Col span={12}>
                            <Form.Item name="maxUses" label="最大使用次数">
                                <InputNumber min={0} precision={0} addonAfter="0 为不限" style={{ width: "100%" }} />
                            </Form.Item>
                        </Col>
                        <Col span={12}>
                            <Form.Item name="enabled" label="状态" valuePropName="checked">
                                <Switch checkedChildren="启用" unCheckedChildren="停用" />
                            </Form.Item>
                        </Col>
                        <Col span={24}>
                            <Form.Item name="remark" label="备注">
                                <Input.TextArea rows={3} />
                            </Form.Item>
                        </Col>
                    </Row>
                </Form>
            </Modal>

            <Modal title="批量生成邀请码" open={batchOpen} width={720} onCancel={() => setBatchOpen(false)} onOk={() => void batchCreate()} okText="生成" cancelText="关闭" confirmLoading={isLoading} destroyOnHidden>
                <Space direction="vertical" size={16} style={{ width: "100%" }}>
                    <Form form={batchForm} layout="vertical" requiredMark={false}>
                        <Row gutter={14}>
                            <Col span={12}>
                                <Form.Item name="type" label="类型" rules={[{ required: true, message: "请选择类型" }]}>
                                    <Select options={typeOptions} />
                                </Form.Item>
                            </Col>
                            <Col span={12}>
                                <Form.Item name="count" label="生成数量" rules={[{ required: true, message: "请输入生成数量" }]}>
                                    <InputNumber min={1} max={200} precision={0} addonAfter="最多 200" style={{ width: "100%" }} />
                                </Form.Item>
                            </Col>
                            <Col span={12}>
                                <Form.Item name="credits" label="兑换额度" rules={currentBatchType === "credits" ? [{ required: true, message: "请输入兑换额度" }] : []}>
                                    <InputNumber min={0} precision={0} disabled={currentBatchType !== "credits"} style={{ width: "100%" }} />
                                </Form.Item>
                            </Col>
                            <Col span={12}>
                                <Form.Item name="maxUses" label="每码最大使用次数">
                                    <InputNumber min={0} precision={0} addonAfter="0 为不限" style={{ width: "100%" }} />
                                </Form.Item>
                            </Col>
                            <Col span={12}>
                                <Form.Item name="enabled" label="状态" valuePropName="checked">
                                    <Switch checkedChildren="启用" unCheckedChildren="停用" />
                                </Form.Item>
                            </Col>
                            <Col span={24}>
                                <Form.Item name="remark" label="备注">
                                    <Input.TextArea rows={2} placeholder="可填写本批次用途" />
                                </Form.Item>
                            </Col>
                        </Row>
                    </Form>
                    {batchResult.length ? (
                        <div className="rounded-md border border-border bg-muted/30 p-3">
                            <div className="mb-3 flex items-center justify-between gap-3">
                                <Typography.Text strong>本次生成 {batchResult.length} 个</Typography.Text>
                                <Button size="small" icon={<CopyOutlined />} onClick={() => copyCodes(batchResult, `已复制 ${batchResult.length} 个邀请码`)}>
                                    复制全部
                                </Button>
                            </div>
                            <Input.TextArea value={batchResult.map((item) => item.code).join("\n")} rows={Math.min(10, Math.max(4, batchResult.length))} readOnly />
                        </div>
                    ) : null}
                </Space>
            </Modal>

            <Modal
                title="删除邀请码"
                open={Boolean(deletingItem)}
                onCancel={() => setDeletingItem(null)}
                onOk={async () => {
                    if (!deletingItem) return;
                    await deleteInviteCode(deletingItem.id);
                    setDeletingItem(null);
                }}
                okText="删除"
                okButtonProps={{ danger: true }}
                cancelText="取消"
            >
                确定删除邀请码「{deletingItem?.code}」吗？
            </Modal>
        </main>
    );
}
